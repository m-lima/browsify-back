docker build -t browsify .
docker stop browsify
docker rm browsify
docker create \
  --name browsify \
  --volume /home/fly/www:/data \
  browsify
