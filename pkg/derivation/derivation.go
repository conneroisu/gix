// Package derivation implements Nix derivation handling
package derivation

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/conneroisu/gix/internal/value"
)

// Derivation represents a Nix derivation.
type Derivation struct {
	// Core fields
	Name      string              `json:"name"`
	Builder   string              `json:"builder"`
	Args      []string            `json:"args"`
	Env       map[string]string   `json:"env"`
	Outputs   map[string]string   `json:"outputs"`
	InputDrvs map[string][]string `json:"inputDrvs"`
	InputSrcs []string            `json:"inputSrcs"`
	System    string              `json:"system"`

	// Computed fields
	Hash      string `json:"hash,omitempty"`
	StorePath string `json:"storePath,omitempty"`
}

// DerivationBuilder helps build derivations.
type DerivationBuilder struct {
	drv *Derivation
}

// NewDerivation creates a new derivation builder.
func NewDerivation(name string) *DerivationBuilder {
	return &DerivationBuilder{
		drv: &Derivation{
			Name:      name,
			Env:       make(map[string]string),
			Outputs:   make(map[string]string),
			InputDrvs: make(map[string][]string),
			InputSrcs: make([]string, 0),
			System:    "x86_64-linux", // default system
		},
	}
}

// SetBuilder sets the builder executable.
func (db *DerivationBuilder) SetBuilder(builder string) *DerivationBuilder {
	db.drv.Builder = builder

	return db
}

// SetArgs sets the builder arguments.
func (db *DerivationBuilder) SetArgs(args []string) *DerivationBuilder {
	db.drv.Args = args

	return db
}

// SetSystem sets the target system.
func (db *DerivationBuilder) SetSystem(system string) *DerivationBuilder {
	db.drv.System = system

	return db
}

// AddEnv adds an environment variable.
func (db *DerivationBuilder) AddEnv(key, value string) *DerivationBuilder {
	db.drv.Env[key] = value

	return db
}

// AddOutput adds an output path.
func (db *DerivationBuilder) AddOutput(name, path string) *DerivationBuilder {
	db.drv.Outputs[name] = path

	return db
}

// AddInputDrv adds an input derivation.
func (db *DerivationBuilder) AddInputDrv(path string, outputs []string) *DerivationBuilder {
	db.drv.InputDrvs[path] = outputs

	return db
}

// AddInputSrc adds an input source.
func (db *DerivationBuilder) AddInputSrc(path string) *DerivationBuilder {
	db.drv.InputSrcs = append(db.drv.InputSrcs, path)

	return db
}

// Build finalizes the derivation.
func (db *DerivationBuilder) Build() *Derivation {
	// Set default output if none specified
	if len(db.drv.Outputs) == 0 {
		db.drv.Outputs["out"] = ""
	}

	// Add name to environment
	db.drv.Env["pname"] = db.drv.Name

	// Compute hash and store path
	db.drv.Hash = db.computeHash()
	db.drv.StorePath = db.computeStorePath()

	// Set output paths if not already set
	for name, path := range db.drv.Outputs {
		if path == "" {
			db.drv.Outputs[name] = filepath.Join(db.drv.StorePath, name)
		}
	}

	return db.drv
}

// computeHash computes the derivation hash.
func (db *DerivationBuilder) computeHash() string {
	// Create a deterministic string representation
	parts := []string{
		"name=" + db.drv.Name,
		"builder=" + db.drv.Builder,
		"args=" + strings.Join(db.drv.Args, ","),
		"system=" + db.drv.System,
	}

	// Add environment variables in sorted order
	var envKeys []string
	for k := range db.drv.Env {
		envKeys = append(envKeys, k)
	}
	sort.Strings(envKeys)

	for _, k := range envKeys {
		parts = append(parts, fmt.Sprintf("env.%s=%s", k, db.drv.Env[k]))
	}

	// Add input derivations in sorted order
	var inputKeys []string
	for k := range db.drv.InputDrvs {
		inputKeys = append(inputKeys, k)
	}
	sort.Strings(inputKeys)

	for _, k := range inputKeys {
		outputs := db.drv.InputDrvs[k]
		sort.Strings(outputs)
		parts = append(parts, fmt.Sprintf("inputDrv.%s=%s", k, strings.Join(outputs, ",")))
	}

	// Add input sources in sorted order
	inputSrcs := make([]string, len(db.drv.InputSrcs))
	copy(inputSrcs, db.drv.InputSrcs)
	sort.Strings(inputSrcs)

	for _, src := range inputSrcs {
		parts = append(parts, "inputSrc="+src)
	}

	// Compute SHA256 hash
	content := strings.Join(parts, "\n")
	hash := sha256.Sum256([]byte(content))

	return hex.EncodeToString(hash[:])[:32] // Use first 32 characters
}

// computeStorePath computes the store path.
func (db *DerivationBuilder) computeStorePath() string {
	return fmt.Sprintf("/nix/store/%s-%s", db.drv.Hash, db.drv.Name)
}

// ToAttrs converts derivation to an attribute set value.
func (d *Derivation) ToAttrs() *value.Attrs {
	attrs := value.NewAttrs()

	// Basic attributes
	attrs.Set("name", value.String(d.Name))
	attrs.Set("builder", value.String(d.Builder))
	attrs.Set("system", value.String(d.System))
	attrs.Set("drvPath", value.String(d.StorePath+".drv"))

	// Args
	argsList := make([]value.Value, len(d.Args))
	for i, arg := range d.Args {
		argsList[i] = value.String(arg)
	}
	attrs.Set("args", value.NewList(argsList...))

	// Outputs
	outAttrs := value.NewAttrs()
	for name, path := range d.Outputs {
		outAttrs.Set(name, value.String(path))
	}
	attrs.Set("outputs", outAttrs)

	// Add individual output attributes
	for name, path := range d.Outputs {
		attrs.Set(name, value.String(path))
	}

	return attrs
}

// FromAttrs creates a derivation from an attribute set.
func FromAttrs(attrs *value.Attrs) (*Derivation, error) {
	// Extract name
	nameVal, ok := attrs.Get("name")
	if !ok {
		return nil, errors.New("derivation missing required 'name' attribute")
	}
	nameStr, ok := nameVal.(value.String)
	if !ok {
		return nil, errors.New("derivation 'name' must be a string")
	}

	// Extract builder
	builderVal, ok := attrs.Get("builder")
	if !ok {
		return nil, errors.New("derivation missing required 'builder' attribute")
	}
	builderStr, ok := builderVal.(value.String)
	if !ok {
		return nil, errors.New("derivation 'builder' must be a string")
	}

	// Create derivation
	db := NewDerivation(string(nameStr))
	db.SetBuilder(string(builderStr))

	// Extract system if present
	if systemVal, ok := attrs.Get("system"); ok {
		if systemStr, ok := systemVal.(value.String); ok {
			db.SetSystem(string(systemStr))
		}
	}

	// Extract args if present
	if argsVal, ok := attrs.Get("args"); ok {
		if argsList, ok := argsVal.(*value.List); ok {
			args := make([]string, argsList.Len())
			for i := 0; i < argsList.Len(); i++ {
				arg := argsList.Get(i)
				if argStr, ok := arg.(value.String); ok {
					args[i] = string(argStr)
				}
			}
			db.SetArgs(args)
		}
	}

	// Extract environment variables
	for _, key := range attrs.Keys() {
		if key == "name" || key == "builder" || key == "system" || key == "args" ||
			key == "outputs" {
			continue
		}
		val, _ := attrs.Get(key)
		if strVal, ok := val.(value.String); ok {
			db.AddEnv(key, string(strVal))
		}
	}

	return db.Build(), nil
}
