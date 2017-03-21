docker pull {{.RepoName}}:{{.Tag}}
docker run -d --name {{.Name}}-backup {{.Params}} --restart=always -e VIRTUAL_HOST={{.Vhost}} -e LETSENCRYPT_HOST={{.Vhost}} {{.RepoName}}:{{.Tag}}
sleep 5
docker stop {{.Name}}
docker rm {{.Name}}
docker run -d --name {{.Name}} {{.Params}} --restart=always -e VIRTUAL_HOST={{.Vhost}} -e LETSENCRYPT_HOST={{.Vhost}} {{.RepoName}}:{{.Tag}}
sleep 5
docker stop {{.Name}}-backup
docker rm {{.Name}}-backup
docker rmi $(docker images -q -f dangling=true)
exit 0
