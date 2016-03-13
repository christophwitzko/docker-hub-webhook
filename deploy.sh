docker pull <%= repo_name %>:<%= tag %>
docker stop <%= name %>
docker rm <%= name %>
docker run -d --name <%= name %> <%= params %> <%= repo_name %>:<%= tag%>
docker rmi $(docker images -q -f dangling=true)
exit 0
