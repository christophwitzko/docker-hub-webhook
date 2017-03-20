docker pull {{.RepoName}}:{{.Tag}}
docker stop {{.Name}}
docker rm {{.Name}}
docker run -d --name {{.Name}} {{.Params}} {{.RepoName}}:{{.Tag}}
docker rmi $(docker images -q -f dangling=true)
exit 0
