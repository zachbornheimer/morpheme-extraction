#include <time.h>
#include <stdlib.h>
#include <stdio.h>

#define DESCENDING 0
#define ASCENDING 1

void quicksort(int **list, int size, int type)
{
	int i = 0, k = 0;
	for (k = 0; k < size; ++k) {
		int pivot = k;
		for (i = 0; i < pivot; ++i) {
			int condition;
			if (type == DESCENDING)
				condition = ((*list)[i] < (*list)[pivot]);
			else
				condition = ((*list)[i] > (*list)[pivot]);
				if (condition) {
					int pivoted = (*list)[pivot];
					(*list)[pivot] = (*list)[i];
					(*list)[i] = pivoted;
				}
			}
		for (i = size-1; i > pivot; --i) {
			int condition;
			if (type == DESCENDING)
				condition = ((*list)[i] > (*list)[pivot]);
			else
				condition = ((*list)[i] < (*list)[pivot]);
			if (condition) {
				int pivoted = (*list)[pivot];
				(*list)[pivot] = (*list)[i];
				(*list)[i] = pivoted;
			}
		}
	}
}

int main()
{
	int size = 10;
	int *list = malloc(sizeof(int) * (size+1));
	int *pivots = malloc(sizeof(int) * (size+1));
	int pivot_index = 0;
	int i;
	srand(time(NULL));
	for (i = size; i >= 0; --i) {
		list[i] = rand()%1000;
	}


	quicksort(&list, size, ASCENDING);
	for (i = 0; i< size; ++i)
		printf("%d,", list[i]);
	printf("\n");
}
