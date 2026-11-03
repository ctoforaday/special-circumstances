#include "llama.h"
#include "json.hpp"

#include <iostream>
#include <fstream>
#include <vector>
#include <string>
#include <chrono>
#include <cmath>
#include <cstring>
#include <algorithm>

static void l2_norm(float * v, int dim) {
    double sum = 0.0;
    for (int i = 0; i < dim; ++i) {
        sum += (double)v[i] * (double)v[i];
    }
    double norm = std::sqrt(sum);
    float scale = norm > 0.0 ? 1.0f / (float)norm : 0.0f;
    for (int i = 0; i < dim; ++i) {
        v[i] *= scale;
    }
}

// Fast self-contained JSON text extractor for {"text": "..."} lines
static std::string extract_text(const std::string & line) {
    size_t k = line.find("\"text\"");
    if (k == std::string::npos) return "";
    size_t colon = line.find(':', k + 6);
    if (colon == std::string::npos) return "";
    size_t q1 = line.find('\"', colon + 1);
    if (q1 == std::string::npos) return "";
    
    std::string s;
    s.reserve(line.size());
    for (size_t i = q1 + 1; i < line.size(); ++i) {
        if (line[i] == '\\' && i + 1 < line.size()) {
            char c = line[i + 1];
            if (c == '\"') { s += '\"'; i++; }
            else if (c == '\\') { s += '\\'; i++; }
            else if (c == 'n') { s += '\n'; i++; }
            else if (c == 'r') { s += '\r'; i++; }
            else if (c == 't') { s += '\t'; i++; }
            else if (c == '/') { s += '/'; i++; }
            else {
                s += '\\';
            }
        } else if (line[i] == '\"') {
            break;
        } else {
            s += line[i];
        }
    }
    return s;
}

int main(int argc, char ** argv) {
    if (argc < 6) {
        std::cerr << "Usage: " << argv[0] << " <model_path> <input_jsonl> <output_bin> <threads> <batch_size_seq> [n_batch_tokens] [flash_attn 0|1]\n";
        return 1;
    }

    std::string model_path = argv[1];
    std::string input_path = argv[2];
    std::string output_path = argv[3];
    int n_threads = std::max(1, std::stoi(argv[4]));
    int batch_size_seq = std::max(1, std::stoi(argv[5]));
    int n_batch_tokens = (argc > 6) ? std::max(512, std::stoi(argv[6])) : std::max(2048, batch_size_seq * 512);
    int flash_attn = (argc > 7) ? std::stoi(argv[7]) : 0;

    // Suppress verbose llama logs to keep stdout clean
    llama_log_set([](enum ggml_log_level level, const char * text, void * user_data) {
        (void)user_data;
        if (level >= GGML_LOG_LEVEL_ERROR || getenv("HARNESS_VERBOSE")) {
            std::cerr << text;
        }
    }, nullptr);

    // Read all texts from input_jsonl
    std::ifstream in(input_path);
    if (!in.is_open()) {
        std::cerr << "Cannot open input: " << input_path << "\n";
        return 1;
    }

    std::vector<std::string> texts;
    std::string line;
    while (std::getline(in, line)) {
        if (line.empty()) continue;
        texts.push_back(nlohmann::json::parse(line).at("text").get<std::string>());
    }
    in.close();

    if (texts.empty()) {
        std::cerr << "No texts loaded from " << input_path << "\n";
        return 1;
    }

    // Measure load time
    auto t0 = std::chrono::steady_clock::now();
    llama_backend_init();

    llama_model_params mparams = llama_model_default_params();
    mparams.progress_callback = [](float progress, void *user_data) {
        (void)progress; (void)user_data;
        return true;
    };
    llama_model * model = llama_model_load_from_file(model_path.c_str(), mparams);
    if (!model) {
        std::cerr << "Failed to load model: " << model_path << "\n";
        return 1;
    }

    llama_context_params cparams = llama_context_default_params();
    cparams.n_ctx = std::max(512 * batch_size_seq, n_batch_tokens);
    cparams.n_batch = n_batch_tokens;
    cparams.n_ubatch = n_batch_tokens; // For non-causal BERT models, n_ubatch == n_batch
    cparams.n_seq_max = batch_size_seq;
    cparams.n_threads = n_threads;
    cparams.n_threads_batch = n_threads;
    cparams.embeddings = true;
    cparams.pooling_type = LLAMA_POOLING_TYPE_MEAN;
    cparams.flash_attn_type = flash_attn ? LLAMA_FLASH_ATTN_TYPE_ENABLED : LLAMA_FLASH_ATTN_TYPE_DISABLED;

    llama_context * ctx = llama_init_from_model(model, cparams);
    if (!ctx) {
        std::cerr << "Failed to init llama_context\n";
        return 1;
    }

    const int dim = llama_model_n_embd(model);
    const llama_vocab * vocab = llama_model_get_vocab(model);
    llama_token sep_token = llama_vocab_sep(vocab);
    if (sep_token == LLAMA_TOKEN_NULL) {
        sep_token = llama_vocab_eos(vocab);
    }

    auto t1 = std::chrono::steady_clock::now();
    double load_secs = std::chrono::duration<double>(t1 - t0).count();

    // Prepare tokenized sequences upfront
    std::vector<std::vector<llama_token>> tokenized_texts(texts.size());
    std::vector<llama_token> tok_buf(2048);
    for (size_t i = 0; i < texts.size(); ++i) {
        int n_tok = llama_tokenize(vocab, texts[i].c_str(), (int32_t)texts[i].length(), tok_buf.data(), (int32_t)tok_buf.size(), true, true);
        if (n_tok < 0) {
            tok_buf.resize(-n_tok + 16);
            n_tok = llama_tokenize(vocab, texts[i].c_str(), (int32_t)texts[i].length(), tok_buf.data(), (int32_t)tok_buf.size(), true, true);
        }
        if (n_tok > 512) {
            tok_buf[511] = sep_token;
            n_tok = 512;
        }
        tokenized_texts[i].assign(tok_buf.begin(), tok_buf.begin() + n_tok);
    }

    // Allocate batch structure with capacity for n_batch_tokens tokens and batch_size_seq sequences
    llama_batch batch = llama_batch_init(n_batch_tokens, 0, batch_size_seq);

    // Embed all texts
    std::vector<float> all_vectors(texts.size() * dim);

    auto t_start_embed = std::chrono::steady_clock::now();

    size_t text_idx = 0;
    while (text_idx < texts.size()) {
        batch.n_tokens = 0;
        int seq_count = 0;
        size_t batch_start_idx = text_idx;

        while (text_idx < texts.size() && seq_count < batch_size_seq) {
            const auto & toks = tokenized_texts[text_idx];
            if (batch.n_tokens + (int)toks.size() > n_batch_tokens && seq_count > 0) {
                // Batch token capacity reached, dispatch current batch
                break;
            }

            int seq_id = seq_count;
            for (size_t pos = 0; pos < toks.size(); ++pos) {
                int cur = batch.n_tokens;
                batch.token[cur] = toks[pos];
                batch.pos[cur] = (llama_pos)pos;
                batch.n_seq_id[cur] = 1;
                batch.seq_id[cur][0] = seq_id;
                batch.logits[cur] = 1; // pool token
                batch.n_tokens++;
            }

            seq_count++;
            text_idx++;
        }

        // Clear context memory before decode
        llama_memory_clear(llama_get_memory(ctx), true);

        int ret = llama_decode(ctx, batch);
        if (ret != 0) {
            std::cerr << "llama_decode failed with ret = " << ret << " at text_idx " << batch_start_idx << "\n";
            return 1;
        }

        for (int s = 0; s < seq_count; ++s) {
            float * emb = llama_get_embeddings_seq(ctx, s);
            if (!emb) {
                emb = llama_get_embeddings(ctx);
            }
            if (!emb) {
                std::cerr << "failed to get embeddings for seq " << s << " at text " << (batch_start_idx + s) << "\n";
                return 1;
            }

            float * dst = all_vectors.data() + (batch_start_idx + s) * dim;
            std::memcpy(dst, emb, dim * sizeof(float));
            l2_norm(dst, dim);
        }
    }

    auto t_end_embed = std::chrono::steady_clock::now();
    double embed_secs = std::chrono::duration<double>(t_end_embed - t_start_embed).count();

    // Write binary output (little-endian float32 vectors)
    if (!output_path.empty() && output_path != "/dev/null") {
        std::ofstream out(output_path, std::ios::binary);
        if (out.is_open()) {
            out.write(reinterpret_cast<const char*>(all_vectors.data()), all_vectors.size() * sizeof(float));
            out.close();
        }
    }

    llama_batch_free(batch);
    llama_free(ctx);
    llama_model_free(model);
    llama_backend_free();

    int n = (int)texts.size();
    double per_s = embed_secs > 0.0 ? (double)n / embed_secs : 0.0;
    double ms_each = n > 0 ? (embed_secs * 1000.0) / (double)n : 0.0;

    // Standard output line matching gteprobe / llamaprobe
    printf("load=%.2fs n=%d workers=%d bseq=%d secs=%.2f per_s=%.1f ms_each=%.1f dim=%d\n",
           load_secs, n, n_threads, batch_size_seq, embed_secs, per_s, ms_each, dim);

    return 0;
}
