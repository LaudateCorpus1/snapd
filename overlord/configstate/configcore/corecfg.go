// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package configcore

import (
	"fmt"
	"os"
	"reflect"

	"github.com/snapcore/snapd/overlord/configstate/config"
)

var (
	Stdout = os.Stdout
	Stderr = os.Stderr
)

// coreCfg returns the configuration value for the core snap.
func coreCfg(tr config.ConfGetter, key string) (result string, err error) {
	var v interface{} = ""
	if err := tr.Get("core", key, &v); err != nil && !config.IsNoOption(err) {
		return "", err
	}
	// TODO: we could have a fully typed approach but at the
	// moment we also always use "" to mean unset as well, this is
	// the smallest change
	return fmt.Sprintf("%v", v), nil
}

// supportedConfigurations contains a set of handled configuration keys.
// The actual values are populated by `init()` functions in each module.
var supportedConfigurations = make(map[string]bool, 32)

func validateBoolFlag(tr config.ConfGetter, flag string) error {
	value, err := coreCfg(tr, flag)
	if err != nil {
		return err
	}
	switch value {
	case "", "true", "false":
		// noop
	default:
		return fmt.Errorf("%s can only be set to 'true' or 'false'", flag)
	}
	return nil
}

// PlainCoreConfig carries a read-only copy of core config and implements
// config.ConfGetter interface.
type PlainCoreConfig map[string]interface{}

// Get implements config.ConfGetter interface.
func (cfg PlainCoreConfig) Get(snapName, key string, result interface{}) error {
	if snapName != "core" {
		return fmt.Errorf("internal error: expected core snap in Get(), %q was requested", snapName)
	}

	val, ok := cfg[key]
	if !ok {
		return &config.NoOptionError{SnapName: snapName, Key: key}
	}

	rv := reflect.ValueOf(result)
	rv.Elem().Set(reflect.ValueOf(val))
	return nil
}

// GetMaybe implements config.ConfGetter interface.
func (cfg PlainCoreConfig) GetMaybe(instanceName, key string, result interface{}) error {
	err := cfg.Get(instanceName, key, result)
	if err != nil && !config.IsNoOption(err) {
		return err
	}
	return nil
}

// fsOnlyContext encapsulates extra options passed to individual core config
// handlers when configuration is applied to a specific root directory with
// FilesystemOnlyApply().
type fsOnlyContext struct {
	RootDir string
}
