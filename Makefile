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

install:
	@go install .