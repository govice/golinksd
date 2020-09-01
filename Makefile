all: 

buildup: 
	docker-compose build && docker-compose up

app: darwin

darwin:
	env GOOS=darwin GOARCH=amd64 go build -o build/darwin/golinksd
	
	- mkdir build/darwin/golinksd.app
	- mkdir build/darwin/golinksd.app/Contents
	- mkdir build/darwin/golinksd.app/Contents/Resources
	- mkdir build/darwin/golinksd.app/Contents/MacOS
	cp assets/golinksd_icon.png build/darwin/golinksd.app/Contents/Resources/golinksd.icns
	cp build/darwin/golinksd build/darwin/golinksd.app/Contents/MacOS/golinksd
	cp assets/Info.plist build/darwin/golinksd.app/Contents/Info.plist

clean:
	rm -rf build/

