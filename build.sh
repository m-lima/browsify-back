docker build -t browsify .
docker stop skull
docker rm browsify
docker create \
  --name browsify \
  --volume /home/fly/www:/data \
  browsify
