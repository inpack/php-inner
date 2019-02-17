[project]
name = php-keeper
version = 0.9.0
vendor = sysinner.com
homepage = http://www.sysinner.com
groups = dev/sys-runtime
description = configuration management tool for php

%build
PREFIX="{{.project__prefix}}"

mkdir -p {{.buildroot}}/bin

CGO_ENABLED=0 GOOS=linux go build -o {{.buildroot}}/bin/php-keeper -a -tags netgo -ldflags '-w -s' main.go

install misc/php-init {{.buildroot}}/bin/php56-init
install misc/php-init {{.buildroot}}/bin/php71-init
install misc/php-init {{.buildroot}}/bin/php72-init
install misc/php-init {{.buildroot}}/bin/php73-init

sed -i 's/{{\.var__pkgname}}/php56/g' {{.buildroot}}/bin/php56-init
sed -i 's/{{\.var__pkgname}}/php71/g' {{.buildroot}}/bin/php71-init
sed -i 's/{{\.var__pkgname}}/php72/g' {{.buildroot}}/bin/php72-init
sed -i 's/{{\.var__pkgname}}/php73/g' {{.buildroot}}/bin/php73-init


%files
LICENSE
misc/

