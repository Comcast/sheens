package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"crypto/sha256"
	"encoding/base64"

	"github.com/Comcast/sheens/core"
	"github.com/Comcast/sheens/crew"

	"github.com/jsccast/yaml"
)

type FileSystemSpecProvider struct {
	sync.RWMutex

	Service *Service
	Dir     string

	timersSpec *core.Spec
	specs      map[string]core.Specter
}

func NewFileSystemSpecProvider(ctx context.Context, s *Service, dir string) (*FileSystemSpecProvider, error) {
	timersSpec := NewTimersSpec()
	if err := timersSpec.Compile(ctx, s.Interpreters, true); err != nil {
		return nil, err
	}
	return &FileSystemSpecProvider{
		Service:    s,
		Dir:        dir,
		specs:      make(map[string]core.Specter, 32),
		timersSpec: timersSpec,
	}, nil
}

func (p *FileSystemSpecProvider) Find(ctx context.Context, s *crew.SpecSource) (core.Specter, error) {
	switch s.Name {
	case "timers":
		return p.timersSpec, nil
	}

	p.RLock()
	spec, have := p.specs[s.Name]
	p.RUnlock()

	if !have {
		return nil, fmt.Errorf(`couldn't find spec named "%s"`, s.Name)
	}
	return spec, nil
}

// ReadSpecs will attempt to gather up MachineSpecs based on YAML
// files in the given directory.
func (p *FileSystemSpecProvider) ReadSpecs(ctx context.Context) error {
	p.Lock()
	defer p.Unlock()

	log.Printf("ReadSpecs %s", p.Dir)

	files, err := ioutil.ReadDir(p.Dir)
	if err != nil {
		return err
	}

	specs := make(map[string]*core.Spec, len(p.specs))

	for _, fi := range files {
		name := fi.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}
		// log.Printf("ReadSpecs loading spec from %s", name)
		bs, err := ioutil.ReadFile(p.Dir + "/" + name)
		if err != nil {
			return err
		}

		var spec core.Spec
		if err = yaml.Unmarshal(bs, &spec); err != nil {
			return err
		}
		// log.Printf("ReadSpecs loaded spec from %s", name)

		// Strip .yaml to get spec name.
		i := strings.LastIndex(name, ".")
		name = name[0:i]
		spec.Name = name

		SetSpecId(&spec)

		// log.Printf("ReadSpecs compiling spec %s", spec.Name)
		if err = spec.Compile(ctx, p.Service.Interpreters, true); err != nil {
			err = fmt.Errorf("%s with '%s'", err, name)
			return err
		}

		core.Logf("Read and compiled %s [%s]", spec.Name, spec.Id)
		specs[name] = &spec
	}
	log.Printf("Loaded %d specs", len(specs))

	specters := make(map[string]core.Specter, len(p.specs))

	for name, spec := range p.specs {
		specters[name] = spec
	}

	for name, spec := range specs {
		specter, have := specters[name]
		if have {
			// log.Printf("updating spec %s", name)
			if updatable, is := specter.(*core.UpdatableSpec); is {
				if err := updatable.SetSpec(spec); err != nil {
					return err
				}
			} else {
				specters[name] = core.NewUpdatableSpec(spec)
			}
		} else {
			// log.Printf("adding spec %s", name)
			specters[name] = core.NewUpdatableSpec(spec)
		}
	}

	p.specs = specters

	return nil
}

// Hash computes the Base64-encoded SHA256 hash of the given data.
func Hash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// SetSpecId generates and sets the Id of the Spec based on a
// canonical (?) JSON representation of Spec.
//
// Warning: Native specs that are identical outside of their native
// actions will get the same id regardless of whether those native
// actions are identical.  So use Spec.Version or Spec.Name to
// differentiate native Specs.
func SetSpecId(s *core.Spec) (string, error) {
	js, err := json.Marshal(&s)
	if err != nil {
		return "", err
	}
	id := Hash(js)
	s.Id = id
	return id, nil
}
