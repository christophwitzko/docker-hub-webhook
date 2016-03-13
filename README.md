# docker-hub-webhook
> Deploy Docker images with [Docker Hub webhooks](https://docs.docker.com/docker-hub/webhooks/).

## Run

    $ docker run -it -p 5000:5000 -e DEFAULT_PARAMS='-e MY_ENV=true' -e DEFAULT_TOKEN=my-token -v /var/run/docker.sock:/var/run/docker.sock:ro christophwitzko/docker-hub-webhook

## Licence

The [MIT License (MIT)](http://opensource.org/licenses/MIT)

Copyright © 2016 [Christoph Witzko](https://twitter.com/christophwitzko)
