listings_version := 0.2

.PHONY: build-bin build-img

build: build-bin build-img

build-bin:
	GOOS=linux GOARCH=amd64 go build -o listings main.go

build-img:
	@docker build -t quay.io/mhausenblas/noteless-listings:$(listings_version) .
	@docker push quay.io/mhausenblas/noteless-listings:$(listings_version)

deploy:
	@kubectl apply -f app.yaml