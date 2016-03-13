FROM mhart/alpine-node:5

RUN apk add --no-cache curl bash && \
    curl -SL https://get.docker.com/builds/Linux/x86_64/docker-latest -o /usr/bin/docker && \
    chmod +x /usr/bin/docker

WORKDIR /src
ADD ["package.json", "deploy.sh", "index.js", "./"]

RUN npm install --production

EXPOSE 5000
CMD ["npm", "start"]
