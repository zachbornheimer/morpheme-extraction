artificial-intelligence-engine
==============================

All my research for Artificial Intelligence
NOTE: The software, if not licensed, will be.  Until that time the following multi-license solution is necessary:

    Individuals may use the software under the terms of the GPLv2 or Later
    Business, Governments, and Militaries  must contact Zysys (zach at zysys.org) before use or looking at the code and must sign a license agreement (the terms are up in the air).

    Terms for licensing are not necessary financial.  Terms may include be a place to test the code, results be recommitted, and/or attribution.

    If licensee must disclose the use of the software/algorithms to Zysys (who are able to sign a non-disclosure agreement) and may never alter the use of the software without another licensing agreement.  This means that if a company desides to use the software for search and now they want to use it for advertising, the company would need to disclose how they would use the software (and how it integrates) and negotiate a new license.

    These terms apply to all of these software without license agreements (including previous and future revisions).

Natural Language Processing: nlp.pl6
===================================
Coded in: Perl 6

Flow:
* Use statistical analysis to discerne the most frequent morphemes in a given scenario [Extraction]
* Use statistical analysis to discerne the context in which those morphemes appear [Extrapolation]
* From those statistics, develop a primative set of rules that outline the morphemes of English.
* From the morphemes, test adding the morphemes to words that appear in thoes scenarios and see if they are spelt correctly.  Learn from that information to determine interesting morpheme rules (like when to use {-es} vs {-s}).
* From that information, be able to give the program words to represent classes of words (ie Verbs, Adjectives, Nouns, etc.)
* Be able to get the program to print out a summarizing report of parts of speech given an input sentence AND be able to have the program change the morphemes in the words (like adding an {-s} or {-es} and similar).


Eventually the nlp.pl6 program can be used to solve the Summarization Problem (given text input, summarize it).  Note, the problem with this program thus far is that the computer does not understand the meaning of words.  At the core, words can be defined in terms of other words, but at the very base level there is sensory association.  Without that sensory association, it cannot define any words.  So, after vision and robotics are complete, this program will be complete.
