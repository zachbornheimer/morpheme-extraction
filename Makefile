CC=cc
DEBUG_CFLAGS=-c -g -Wall
CFLAGS=-c -Wall -O
LDFLAGS=
SOURCES=functions.c file.c directory.c word_t.c alphabet.c ngram_t.c nlp.c morpheme_t.c morpheme_list_t.c
OBJECTS=$(SOURCES:.c=.o)
DEBUG_OBJECTS=$(SOURCES:.c=.debug.o)
DEBUG_EXE=nlp.debug
PRODUCTION_EXE=nlp

debug: EXECUTABLE=$(DEBUG_EXE)
debug: CFLAGS=$(DEBUG_CFLAGS)
debug: OBJECTS=$(DEBUG_OBJECTS)
debug: $(SOURCES) $(DEBUG_EXE)

optimized: EXECUTABLE=$(PRODUCTION_EXE)
optimized: $(SOURCES) $(PRODUCTION_EXE)

all: debug optimized

$(PRODUCTION_EXE): $(OBJECTS)
	$(CC) $(LDFLAGS) $(OBJECTS) -o $@

$(DEBUG_EXE): $(DEBUG_OBJECTS)
	$(CC) $(LDFLAGS) $(DEBUG_OBJECTS) -o $@

%.debug.o: %.c
	$(CC) $(CFLAGS) $< -o $@

%.o: %.c
	$(CC) $(CFLAGS) $< -o $@

clean:
	rm -rf *.o $(PRODUCTION_EXE) $(DEBUG_EXE)
