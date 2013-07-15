#!/usr/bin/perl6


use Zysys::Create::MeaningMap;
use v6;
:autoflush;

class NLPEngine is NLPClass_MM {
    has $.corpusDir;
    has $.zedramOutput is rw;
    has $.humanity;
    has $.explicitDeclarations;
    has $!completedAnalysis;

    method analyze() {
        return !$!completedAnalysis;
    }




}
