# in development
make:
	@go install .   ;
	@sh -c '	   	\
		cd tests;  	\
		mug -gen; 	\
	';

install:
	@go install .