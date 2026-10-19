// Package secretstest draws keys in real vendor formats for the secrets gate's tests.
//
// Every key is drawn at test time over the vendor's whole alphabet, from a seed, so the
// source holds no scannable value: the gate refuses any Bash payload carrying one, including
// the payload that would write a test of the gate, and a checked-in key in a real shape is what
// push protection flags. Every drawn value is random filler in a real shape; none is a live
// credential.
package secretstest

import (
	"hash/fnv"
	"math/rand"
	"strings"
)

// Alphabets, as the vendors publish them.
const (
	Upper     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lower     = "abcdefghijklmnopqrstuvwxyz"
	Digit     = "0123456789"
	Base62    = Upper + Lower + Digit
	Base64URL = Base62 + "_-"
	Base64    = Base62 + "+/"
	Base32    = Upper + "234567"
	HexLower  = "0123456789abcdef"
	UpperNum  = Upper + Digit
	LowerNum  = Lower + Digit
)

// S draws n characters from alpha.
func S(r *rand.Rand, alpha string, n int) string {
	var b strings.Builder
	for range n {
		b.WriteByte(alpha[r.Intn(len(alpha))])
	}
	return b.String()
}

// Dashes is the five-dash run of PEM armor.
var Dashes = strings.Repeat("-", 5)

// Lines draws n lines of 64 base64 characters, joined by newlines.
func Lines(r *rand.Rand, n int) string {
	l := make([]string, n)
	for i := range l {
		l[i] = S(r, Base64, 64)
	}
	return strings.Join(l, "\n")
}

// Format is one vendor key format and the gitleaks rule id that refuses it.
type Format struct {
	Name string
	Rule string
	Draw func(r *rand.Rand) string
}

// Vendors are formats gitleaks' default configuration refuses, one or more per vendor.
var Vendors = []Format{
	{"aws access key id", "aws-access-token", func(r *rand.Rand) string { return "AK" + "IA" + S(r, Base32, 16) }},
	{"aws sts key id", "aws-access-token", func(r *rand.Rand) string { return "AS" + "IA" + S(r, Base32, 16) }},
	{"github classic pat", "github-pat", func(r *rand.Rand) string { return "ghp" + "_" + S(r, Base62, 36) }},
	{"github fine-grained pat", "github-fine-grained-pat", func(r *rand.Rand) string {
		return "github" + "_pat_" + S(r, Base62, 22) + "_" + S(r, Base62, 59)
	}},
	{"slack bot token", "slack-bot-token", func(r *rand.Rand) string {
		return "xo" + "xb-" + S(r, Digit, 13) + "-" + S(r, Digit, 13) + "-" + S(r, Base62, 24)
	}},
	{"slack webhook", "slack-webhook-url", func(r *rand.Rand) string {
		return "https://hooks." + "slack.com/services/T" + S(r, UpperNum, 8) + "/B" + S(r, UpperNum, 10) + "/" + S(r, Base62, 24)
	}},
	{"stripe live key", "stripe-access-token", func(r *rand.Rand) string { return "sk" + "_live_" + S(r, Base62, 24) }},
	{"gitlab pat", "gitlab-pat", func(r *rand.Rand) string { return "gl" + "pat-" + S(r, Base64URL, 20) }},
	{"npm token", "npm-access-token", func(r *rand.Rand) string { return "np" + "m_" + S(r, LowerNum, 36) }},
	{"openai project key", "openai-api-key", func(r *rand.Rand) string {
		return "sk-" + "proj-" + S(r, Base64URL, 74) + "T3Blbk" + "FJ" + S(r, Base64URL, 73) + S(r, Base62, 1)
	}},
	{"anthropic api key", "anthropic-api-key", func(r *rand.Rand) string {
		return "sk-" + "ant-api03-" + S(r, Base64URL, 93) + "AA"
	}},
	{"rsa private key", "private-key", func(r *rand.Rand) string {
		return Dashes + "BEGIN RSA PRIVATE KEY" + Dashes + "\n" + Lines(r, 3) + "\n" + Dashes + "END RSA PRIVATE KEY" + Dashes
	}},
}

// Rand is a source seeded by name, so each format's draws are the same on every run.
func Rand(name string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(name))
	return rand.New(rand.NewSource(int64(h.Sum64())))
}

// Key draws one key of the named Vendors format.
func Key(name string) string {
	for _, f := range Vendors {
		if f.Name == name {
			return f.Draw(Rand(name))
		}
	}
	panic("secretstest: no format " + name)
}
