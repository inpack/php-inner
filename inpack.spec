[project]
name = php-inner
version = 0.1.1
vendor = sysinner.com
homepage = http://www.sysinner.com
groups = dev/sys-runtime
description = configuration management tool for php

%build
PREFIX="{{.project__prefix}}"

mkdir -p {{.buildroot}}/bin

CGO_ENABLED=0 GOOS=linux go build -o {{.buildroot}}/bin/php-inner -a -tags netgo -ldflags '-w -s' inner.go
install misc/php-inner-init {{.buildroot}}/bin/php-inner-init
sed -i 's/{{\.var__pkgname}}/php71/g' {{.buildroot}}/bin/php-inner-init


%files
LICENSE
misc/

