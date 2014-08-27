SHELL := /bin/bash
TARGETS = solrcount

imports:
	goimports -w .

fmt:
	go fmt ./...

vet:
	go vet ./...

all: fmt test
	go build

install:
	go install

clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -f solrcount-*.x86_64.rpm
	rm -f debian/solrcount*.deb
	rm -rf debian/solrcount/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

solrcount:
	go build cmd/solrcount/solrcount.go

# ==== packaging

deb: $(TARGETS)
	mkdir -p debian/solrcount/usr/sbin
	cp $(TARGETS) debian/solrcount/usr/sbin
	cd debian && fakeroot dpkg-deb --build solrcount .

REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm-compatible
	cp solrcount-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/solrcount.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh solrcount
	cp $(HOME)/rpmbuild/RPMS/x86_64/solrcount*.rpm .

# ==== vm-based packaging

PORT = 2222
SSHCMD = ssh -o StrictHostKeyChecking=no -i vagrant.key vagrant@127.0.0.1 -p $(PORT)
SCPCMD = scp -o port=$(PORT) -o StrictHostKeyChecking=no -i vagrant.key

VMHOME = /home/vagrant
OWNER = github.com/miku

# Helper to build RPM on a RHEL6 VM, to link against glibc 2.12
vagrant.key:
	curl -sL "https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant" > vagrant.key
	chmod 0600 vagrant.key

setup: vagrant.key
	$(SSHCMD) "sudo yum install -y http://ftp.riken.jp/Linux/fedora/epel/6/i386/epel-release-6-8.noarch.rpm"
	$(SSHCMD) "sudo yum install -y golang git rpm-build"
	$(SSHCMD) "mkdir -p $(VMHOME)/src/$(OWNER)"
	$(SSHCMD) "cd $(VMHOME)/src/$(OWNER) && git clone https://$(OWNER)/solrcount.git"

rpm-compatible: vagrant.key
	$(SSHCMD) "cd $(VMHOME)/src/$(OWNER)/solrcount && GOPATH=$(VMHOME) go get ./..."
	$(SSHCMD) "cd $(VMHOME)/src/$(OWNER)/solrcount && git pull origin master && pwd && GOPATH=$(VMHOME) make rpm"
	$(SCPCMD) vagrant@127.0.0.1:$(VMHOME)/src/$(OWNER)/solrcount/*rpm .

