build:
	CC=/usr/bin/musl-gcc go build --ldflags '-linkmode external -extldflags "-static"'

run-cron:
	go build
	./gogitmirror cron

upload: build
	scp gogitmirror mscom:/home/janitor/gogitmirror_tmp_upload
	ssh -t mscom "sudo mv -f /home/janitor/gogitmirror_tmp_upload /usr/scripts/common/gogitmirror && cd /usr/scripts/common/ && sudo quickpush gogitmirror-auto-upload"
