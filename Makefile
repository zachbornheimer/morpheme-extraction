CC=cc
CFLAGS=-c -g -Wall
LDFLAGS=
SOURCES=functions.c file.c directory.c word_t.c alphabet.c ngram_t.c nlp.c
OBJECTS=$(SOURCES:.c=.o)
EXECUTABLE=nlp

all: $(SOURCES) $(EXECUTABLE)

$(EXECUTABLE): $(OBJECTS)
	$(CC) $(LDFLAGS) $(OBJECTS) -o $@

.c.o:
	$(CC) $(CFLAGS) $< -o $@

clean:
	rm -rf *o $(EXECUTABLE)
