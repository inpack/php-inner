// Copyright 2017 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/hooto/hflag4g/hflag"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/sysinner/incore/inapi"
)

type mod struct {
	name     string
	priority int
}

func mod_get(name string) int {
	for _, v := range php_mods {
		if v.name == name {
			return v.priority
		}
	}
	return 0
}

var (
	pod_inst    = "/home/action/.sysinner/pod_instance.json"
	pod_inagent = "/home/action/.sysinner/inagent"
	php_prefix  = "/home/action/apps/%s"
	php_ini     = "/home/action/apps/%s/etc/php.ini"
	php_modini  = "/home/action/apps/%s/etc/php.d/%s.ini"
	php_fpmcfg  = "/home/action/apps/%s/etc/php-fpm.conf"
	php_fpmwww  = "/home/action/apps/%s/etc/php-fpm.d/www.conf"
	php_rels    = types.ArrayString([]string{"php56", "php71", "php72"})
	php_rel     = "php71"
	php_mods    = []mod{
		{"opcache", 10},
		{"bcmath", 20},
		{"bz2", 20},
		{"ctype", 20},
		{"curl", 20},
		{"exif", 20},
		{"ftp", 20},
		{"gd", 20},
		{"gettext", 20},
		{"gmp", 20},
		{"iconv", 20},
		{"intl", 20},
		{"json", 20},
		{"mbstring", 20},
		{"mcrypt", 20},
		{"mysqlnd", 20},
		{"pgsql", 20},
		{"pspell", 20},
		{"simplexml", 20},
		{"soap", 20},
		{"sockets", 20},
		{"sqlite3", 20},
		{"tokenizer", 20},
		{"xml", 20},
		{"xsl", 20},
		{"zip", 20},
		{"mysqli", 30},
		{"pdo", 30},
		{"pdo_mysql", 30},
		{"pdo_pgsql", 30},
		{"pdo_sqlite", 30},
		{"wddx", 30},
		{"xmlrpc", 30},
	}
)

func main() {
	if err := pod_init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if v, ok := hflag.ValueOK("php-rel"); ok && !php_rels.Has(v.String()) {
		fmt.Println("invalid php-rel")
		os.Exit(1)
	}

	if _, ok := hflag.ValueOK("php-init"); ok {
		if err := base_set(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if v, ok := hflag.ValueOK("php-modules"); ok {
		if err := module_sets(v.String()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if _, ok := hflag.ValueOK("php-fpm-on"); ok {
		if err := fpm_on(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func base_set() error {

	php_ini_path := fmt.Sprintf(php_ini, php_rel)

	fp, err := os.Open(php_ini_path + ".default")
	if err != nil {
		return err
	}
	defer fp.Close()

	src, err := ioutil.ReadAll(fp)
	if err != nil {
		return err
	}

	sets := map[string]string{
		"session__save_path": "/home/action/var/tmp",
	}
	//
	tpl, err := template.New("s").Parse(string(src))
	if err != nil {
		return err
	}

	var dst bytes.Buffer
	if err := tpl.Execute(&dst, sets); err != nil {
		return err
	}

	fpdst, err := os.OpenFile(php_ini_path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fpdst.Close()

	fpdst.Seek(0, 0)
	fpdst.Truncate(0)

	_, err = fpdst.Write(dst.Bytes())

	return err
}

func fpm_on() error {

	cfgs := []string{
		fmt.Sprintf(php_fpmcfg, php_rel),
		fmt.Sprintf(php_fpmwww, php_rel),
	}

	sets := map[string]string{}

	for _, v := range cfgs {

		fp, err := os.Open(v + ".default")
		if err != nil {
			return err
		}
		defer fp.Close()

		src, err := ioutil.ReadAll(fp)
		if err != nil {
			return err
		}
		//
		tpl, err := template.New("s").Parse(string(src))
		if err != nil {
			return err
		}

		var dst bytes.Buffer
		if err := tpl.Execute(&dst, sets); err != nil {
			return err
		}

		fpdst, err := os.OpenFile(v, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer fpdst.Close()

		fpdst.Seek(0, 0)
		fpdst.Truncate(0)

		_, err = fpdst.Write(dst.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func module_sets(v string) error {

	vs := types.ArrayString(strings.Split(v, ","))

	for _, m := range vs {
		if strings.HasPrefix(m, "pdo_") {
			switch m {
			case "pdo_mysql":
				vs.Set("mysqlnd")

			case "pdo_pgsql":
				vs.Set("pgsql")

			case "pdo_sqlite":
				vs.Set("sqlite3")
			}
			vs.Set("pdo")
		}
	}

	if v == "all" {
		for _, m := range php_mods {
			vs.Set(m.name)
		}
	}

	for _, m := range vs {
		if priority := mod_get(m); priority > 0 {
			mod_body := fmt.Sprintf("extension=%s.so\n", m)
			if err := module_set_file(fmt.Sprintf("%d-%s", priority, m), mod_body); err != nil {
				return err
			}
		}
	}

	return nil
}

func module_set_file(name string, s string) error {

	fp, err := os.OpenFile(fmt.Sprintf(php_modini, php_rel, name), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	fp.Seek(0, 0)
	fp.Truncate(0)
	_, err = fp.Write([]byte(s))

	return err
}

func pod_init() error {

	var inst inapi.Pod
	if err := json.DecodeFile(pod_inst, &inst); err != nil {
		return err
	}

	if inst.Spec == nil ||
		inst.Spec.Box.Resources == nil {
		return errors.New("Not Pod Instance Setup")
	}

	for _, app := range inst.Apps {
		for _, p := range app.Spec.Packages {
			if php_rels.Has(p.Name) {
				php_rel = p.Name
				break
			}
		}
	}

	_, err := os.Stat(fmt.Sprintf(php_prefix+"/bin/php", php_rel))
	return err
}
