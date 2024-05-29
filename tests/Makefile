clean:
	podman stop comics_db_test_c || true
	podman rm comics_db_test_c || true
	podman rmi localhost/comics_db_test:latest || true
	docker stop comics_db_test_c || true
	docker rm comics_db_test_c || true
	docker rmi comics_db_test || true
