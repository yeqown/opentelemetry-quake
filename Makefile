tag:
	git tag ${TAG} -m "Release ${TAG}"
	@ echo "Remember to push the tag to the remote repo: make release-tag"

tag-contrib-grpc:
	- git tag contrib/grpc/${TAG} -m "Release ${TAG}"

tag-contrib-gin:
	- git tag contrib/gin/${TAG} -m "Release ${TAG}"

tag-contrib-resty:
	- git tag contrib/resty/${TAG} -m "Release ${TAG}"

tag-contrib-all: tag-contrib-gin tag-contrib-grpc tag-contrib-resty
	@ echo "All contrib repos have been tagged"

release-tag:
	git push origin --tags
