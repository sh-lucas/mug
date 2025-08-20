# in development
make:
	@go install ./cmd/mug/   ;
	@sh -c '	   	\
		cd tests;  	\
		mug -gen; 	\
	';

install:
	@go install .