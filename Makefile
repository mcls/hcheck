default: install
	hcheck https://www.google.com https://www.facebook.com https://github.com

install:
	go install
