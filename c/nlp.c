/*
 * nlp.c
 * Written by Z. Bornheimer (Zysys)
 * Designed to produce perl6 code for implementation
 * Follows the Design Spec written by Z. Bornheimer
 *
 * Part of the Computerized Human Language Acquisition System
 */

#include "constants.h"
#include "prototypes.h"
#include "structures.h"
#include "functions.h"

int main() {
    assemble_structures();
    alphabet_identification();
    morpheme_extraction();
    lexical_category_identification();
    syntax_model_creation();
    meaning_map_generation();

    return 0;
}
