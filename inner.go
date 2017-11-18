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

var (
	pod_inst    = "/home/action/.sysinner/pod_instance.json"
	pod_inagent = "/home/action/.sysinner/inagent"
	php_prefix  = "/home/action/apps/%s"
	php_ini     = "/home/action/apps/%s/etc/php.ini"
	php_modini  = "/home/action/apps/%s/etc/php.d/modules.ini"
	php_fpmcfg  = "/home/action/apps/%s/etc/php-fpm.conf"
	php_fpmwww  = "/home/action/apps/%s/etc/php-fpm.d/www.conf"
	php_defs    = types.ArrayString([]string{"php56", "php71", "php72"})
	php_def     = "php71"
	php_mods    = types.ArrayString([]string{
		"bcmath",
		"bz2",
		"ctype",
		"curl",
		"exif",
		"ftp",
		"gd",
		"gettext",
		"gmp",
		"iconv",
		"intl",
		"json",
		"mbstring",
		"mcrypt",
		"mysqli",
		"mysqlnd",
		"opcache",
		"pdo",
		"pdo_mysql",
		"pdo_pgsql",
		"pdo_sqlite",
		"pgsql",
		"pspell",
		"simplexml",
		"soap",
		"sockets",
		"sqlite3",
		"tokenizer",
		"wddx",
		"xmlrpc",
		"xml",
		"xsl",
		"zip",
	})
)

func main() {
	if err := pod_init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, ok := hflag.Value("php"); ok {
		if err := base_set(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if v, ok := hflag.Value("php_modules"); ok {
		if err := module_sets(v.String()); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if _, ok := hflag.Value("php_fpm"); ok {
		if err := fpm_on(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func base_set() error {

	php_ini_path := fmt.Sprintf(php_ini, php_def)

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
		fmt.Sprintf(php_fpmcfg, php_def),
		fmt.Sprintf(php_fpmwww, php_def),
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
	modules := ""

	var (
		vs  = strings.Split(v, ",")
		vs2 types.ArrayString
	)
	for _, m := range vs {
		if php_mods.Has(m) {
			if strings.HasPrefix(m, "pdo_") {
				vs2.Set("pdo")
			}
			vs2.Set(m)
		}
	}

	for _, m := range vs2 {
		modules += fmt.Sprintf("extension=%s.so\n", m)
	}

	if len(modules) > 0 {

		fp, err := os.OpenFile(fmt.Sprintf(php_modini, php_def), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}

		fp.Seek(0, 0)
		fp.Truncate(0)
		_, err = fp.Write([]byte(modules))
		if err != nil {
			return err
		}
	}

	return nil
}

func pod_init() error {

	var inst inapi.Pod
	if err := json.DecodeFile(pod_inst, &inst); err != nil {
		return err
	}

	if inst.Spec == nil ||
		len(inst.Spec.Boxes) == 0 ||
		inst.Spec.Boxes[0].Resources == nil {
		return errors.New("Not Pod Instance Setup")
	}

	for _, app := range inst.Apps {
		for _, p := range app.Spec.Packages {
			if php_defs.Has(p.Name) {
				php_def = p.Name
				break
			}
		}
	}

	_, err := os.Stat(fmt.Sprintf(php_prefix+"/bin/php", php_def))
	return err
}
