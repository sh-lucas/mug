# in development
make:
	@go install .   ;
	@sh -c '	   	\
		cd tests;  	\
		mug --gen; 	\
	';
	@sh -c '			\
		cd tests	; \
		go run .	; \
	';

install:
	@go install .