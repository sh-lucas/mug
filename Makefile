# in development
make:
	@go install .   ;
	@sh -c '	   \
		cd tests;  \
		mug; 	   \
	';

# watcher test
watch:
	@go install .   ;
	@sh -c '       \
		cd tests;  \
		mug watch  \
	';