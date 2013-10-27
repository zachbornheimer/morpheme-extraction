CC=gcc
CFLAGS=-c -Wall
LDFLAGS=
SOURCES=functions.c file.c directory.c words_t.c alphabet.c 
OBJECTS=$(SOURCES:.c=.o)
EXECUTABLE=alphabet

all: $(SOURCES) $(EXECUTABLE)

$(EXECUTABLE): $(OBJECTS)
	$(CC) $(LDFLAGS) $(OBJECTS) -o $@

.c.o:
	$(CC) $(CFLAGS) $< -o $@

clean:
	rm -rf *o $(EXECUTABLE)
