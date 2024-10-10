package profile

import (
	"errors"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/dtbead/moonpool/config"
)

type Profile struct {
	cpu, mem *os.File
}

func New(profilerType string) (Profile, error) {
	if err := os.MkdirAll("profile", 0777); err != nil && !errors.Is(err, os.ErrExist) {
		return Profile{}, err
	}

	fileCPU, err := os.Create("profile/cpu.prof")
	if err != nil {
		return Profile{}, err
	}

	fileMem, err := os.Create("profile/mem.prof")
	if err != nil {
		fileCPU.Close()
		return Profile{}, err
	}

	switch profilerType {
	case config.PROFILING_CPU:
		pprof.StartCPUProfile(fileCPU)
	default:
		fileCPU.Close()
		fileMem.Close()
		return Profile{}, errors.New("unknown profiler type")
	}

	return Profile{cpu: fileCPU, mem: fileMem}, nil
}

func (p Profile) Stop() error {
	pprof.StopCPUProfile()
	p.cpu.Close()

	runtime.GC()
	if err := pprof.WriteHeapProfile(p.mem); err != nil {
		return err
	}

	return nil
}
