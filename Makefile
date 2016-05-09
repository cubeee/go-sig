all: build-gb

deps:
	./install_dependencies.sh

build-gb:
	gb build

package: deps build-gb
	sudo rm -rf build/
	mkdir -p build/opt/gosig/images
	mkdir -p build/etc/systemd/system
	cp bin/signature build/opt/gosig/gosig
	cp -r assets build/opt/gosig/
	cp -r public build/opt/gosig/
	cp -r templates build/opt/gosig/
	cp systemd/go-sig.service build/etc/systemd/system/go-sig.service
	sudo chown -R gosig: build/opt
	sudo chown -R root: build/etc
	echo 2.0 > build/debian-binary
	echo "Package: go-sig" > build/control
	echo "Version: 1.0" >> build/control
	echo "Architecture: all" >> build/control
	echo "Section: net" >> build/control
	echo "Maintainer: cubeee <cubeee.gh@gmail.com>" >> build/control
	echo "Priority: optional" >> build/control
	echo "Homepage: https://sig.scapelog.com/"
	echo "Description: Dynamically generated and updated skill goal signatures for RuneScape players" >> build/control
	tar cvzf build/data.tar.gz -C build etc opt
	tar cvzf build/control.tar.gz -C build control
	cd build && ar rc go-sig.deb debian-binary control.tar.gz data.tar.gz && cd ..
