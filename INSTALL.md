Installation instructions for SPP server
========================================

This installation manual covers everything for you to get the SPP service up and running.

The procedure consists of five parts:

1. Installing the SPP web application
2. Installing nginx
3. Installing PostgreSQL
4. Creating a supervisord config for the application
5. Downloading contraction hiearchies graph files

Anything with a # is a comment and anything after that should be ignored.

1. SPP Web App
--------------

1. Checkout the SPP files and build the server:

	svn co https://svn.cc.jyu.fi/srv/svn/transopt/products/ch/trunk
	cd trunk/app
	go get github.com/hoisie/redis
	go get github.com/hoisie/web
	go get github.com/bmizerany/pq
	go build -ldflags="-r ."

2. server should just fine if you run ./server, but requests WILL FAIL because geoindexing won't work and nginx is missing!

2. Configure Nginx as a reverse proxy
-----------------

1. Install nginx with apt-get or yum
2. Edit the nginx config:
	sudo vim /etc/nginx/nginx.conf
3. Find the line "server {"
4. Add the following underneath:
        location / {
                proxy_pass      http://127.0.0.1:10080;
        } 
5. Save. Nginx is now configured as a reverse proxy that routes everything from port 80 to port 10080, where the SPP web app listens. This setting can be reconfigured in production.json in the app directory.

PostgreSQL
----------

1. Install postgresql with apt-get or yum
2. Import the tables and databases

Installing Go
-------------

Either install it with apt-get or yum, but be certain it's go 1.1! This won't compile with Go 1.0.

Here's how to compile it from source

1. Download the source
	wget https://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz
2. Untar
	tar xaf go1.1.2.linux-amd64.tar.gz
	cd go/src
	./all.bash
3. Once it prints "ALL TESTS PASSED", do the following:

1. Configure GO
	Open ~/.bash_profile with an editor, e.g. "nano ~/.bash_profile"
	Add the following lines:
	export GOPATH=$HOME/gocode
	export GOROOT=$HOME/go
	(This sets important environment variables for the Go runtime.)
	Modify PATH to be:
	PATH=$PATH:$HOME/bin:$HOME/go/bin:$HOME/gocode/bin
	(This makes sure the go runtime libraries are in PATH.)
	Load the new configuration thus:
	source ~/.bash_profile
