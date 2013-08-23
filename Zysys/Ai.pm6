#!/usr/bin/perl6
use v6;

module Zysys::Ai;

# This is part of a series of modules that give an OO API for this design
#
# There exists an implementation spec in my codebook (physical) and also
#   a design spec.  That design spec is what was delievered for the
#   provisional patent application.  That document is entitled:
#   "Computerized Language Acquistion System and Associated Processes"
#
# This system is covered under patent protection.
#
# While I don't like that I need patent protection, it is necessary
#   until I can fully understand the implications here.  I don't want
#   this system being misused to harm humankind.
#
# If you are planning on not heading my warning about this and plan
#   on just implementing this code or reworking this, don't.
#   You may make a huge mistake that could lead to Aritifical Intelligence
#   of Hollwood's design.  The implications of this project are profound.
#   Please contact me if you decide to do something with this as I am
#   working on discovering the implications of this software and the
#   concepts behind it.
#
# This section was written August 22, 2013 by Z. Bornheimer.
#
# Please understand the concepts behind this software before modifying it.
# 
# Also, follow the style guide:
#   http://zysys.org/wiki/index.php/Perl6_Style_Guide


#########################################################################
# Module Design:                                                        #
#   Alphabet Identification -> Zysys::Identify::Alphabet                #
#   Lexical Category Identification -> Zysys::Identify::LexicalCategory #
#   Morpheme Extraction -> Zysys::Extract::Morpheme                     #
#   Syntax Model Creation -> Zysys::Create::SyntaxModel                 #
#   Meaning Map Creation -> Zysys::Create::MeaningMap                   #
#########################################################################


use Zysys::Identify::Alphabet;
use Zysys::Identify::LexicalCategories;
use Zysys::Extract::Morphemes;
use Zysys::Create::SyntaxMode;
use Zysys::Create::MeaningMap;


:autoflush;

class NLPEngine is NLPClass_MM {
    has $.corpusDir;
    has $.zedramOutput is rw;
    has $.humanity;
    has $.explicitDeclarations;
    has $!completedAnalysis;

    method analyze() {
        if !$!explicitDeclaration {
            this.assemble_structures();
            this.alphabet_identification();
            this.morpheme_extraction();
            this.lexical_category_identification();
            this.syntax_model_creation();
            this.meaning_map_generation();
        }
        return !$!completedAnalysis;
    }

    method assemble_structures() {

    }

    method alphabet_identification() {

    }

    method morpheme_extraction() {

    }

    method lexical_category_identification() {
    
    }

    method syntax_model_creation() {

    }

    method meaning_map_generation() {

    }

}
