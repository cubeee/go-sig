build:
	gb build signature

package: build
	sudo rm -rf build/
	mkdir -p build/opt/gosig/images
	mkdir -p build/etc/supervisor/conf.d
	cp bin/signature build/opt/gosig/gosig
	cp base.png build/opt/gosig/
	cp MuseoSans_500.ttf build/opt/gosig
	cp supervisor/go-sig.conf build/etc/supervisor/conf.d/go-sig.conf
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
	echo "Description: Dynamically generated skill goal signatures for RuneScape players" >> build/control
	tar cvzf build/data.tar.gz -C build etc opt
	tar cvzf build/control.tar.gz -C build control
	cd build && ar rc go-sig.deb debian-binary control.tar.gz data.tar.gz && cd ..
