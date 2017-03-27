all:
	/usr/local/go/bin/go build bot.go TravianClient.go TravianData.go Settings.go
run:
	/usr/local/go/bin/go run bot.go TravianClient.go TravianData.go Settings.go

# go build -ldflags "-s -w"