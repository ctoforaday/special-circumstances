package vecembed

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"golang.org/x/sys/cpu"
)

// cpuDesc names the processor the measurement ran on and the instruction-set extensions ONNX
// Runtime's kernels dispatch on, for the S5 and S6 lines: two runners of one image can differ in
// both, and the numbers move with them. Best-effort: a value it cannot read is "unknown", never a
// failure.
func cpuDesc() string {
	return fmt.Sprintf("cpu=%q isa=%s", cpuModel(), cpuISA())
}

func cpuModel() string {
	switch runtime.GOOS {
	case "linux":
		f, err := os.Open("/proc/cpuinfo")
		if err != nil {
			return "unknown"
		}
		defer f.Close()
		// x86 names its model; arm64 names its implementer and part.
		var impl, part string
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			k, v, ok := strings.Cut(sc.Text(), ":")
			if !ok {
				continue
			}
			switch k, v = strings.TrimSpace(k), strings.TrimSpace(v); k {
			case "model name":
				return v
			case "CPU implementer":
				impl = v
			case "CPU part":
				part = v
			}
		}
		if impl != "" || part != "" {
			return "implementer " + impl + " part " + part
		}
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	case "windows":
		if v := os.Getenv("PROCESSOR_IDENTIFIER"); v != "" {
			return v
		}
	}
	return "unknown"
}

func cpuISA() string {
	var have []string
	add := func(name string, ok bool) {
		if ok {
			have = append(have, name)
		}
	}
	switch runtime.GOARCH {
	case "amd64":
		add("avx", cpu.X86.HasAVX)
		add("avx2", cpu.X86.HasAVX2)
		add("fma", cpu.X86.HasFMA)
		add("avxvnni", cpu.X86.HasAVXVNNI)
		add("avx512f", cpu.X86.HasAVX512F)
		add("avx512bw", cpu.X86.HasAVX512BW)
		add("avx512vnni", cpu.X86.HasAVX512VNNI)
	case "arm64":
		add("asimd", cpu.ARM64.HasASIMD)
		add("dotprod", cpu.ARM64.HasASIMDDP)
		add("i8mm", cpu.ARM64.HasI8MM)
		add("sve", cpu.ARM64.HasSVE)
		add("sve2", cpu.ARM64.HasSVE2)
	}
	if len(have) == 0 {
		return "unknown"
	}
	return strings.Join(have, ",")
}
