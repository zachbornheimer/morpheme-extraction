#define V_PRINT(msg) \
	if (verbose_mode == ON) \
		printf("%s\n", msg);

#define ASSERT(condition, error_code) \
	if (condition == 0) { \
		printf("Assertion: '%s' failed.\n", #condition); \
		exit(error_code); \
	}

#define REALLOC_CHECK(arg) \
	ASSERT((arg != NULL), E_REALLOC);

