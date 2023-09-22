docker:
	docker buildx build --no-cache -t germanorizzo/seif:v0.0.2 .
	docker tag germanorizzo/seif:v0.0.2 germanorizzo/seif:latest
	docker push germanorizzo/seif:v0.0.2
	docker push germanorizzo/seif:latest

run:
	npm run build
	docker run -i --rm --name seif -v `pwd`:/data -p 12321:12321 germanorizzo/sqliterg:latest /sqliterg --db /data/seif.db --serve-dir /data/dist --index-file index.html