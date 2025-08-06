# in development
make:
	@go install .   ;
	@sh -c '	   	\
		cd tests;  	\
		mug; 	   		\
	';
	@sh -c '			\
		cd tests	; \
		go run .	; \
	';

# watcher test
watch:
	@go install .   ;
	@sh -c '       \
		cd tests;  \
		mug watch  \
	';