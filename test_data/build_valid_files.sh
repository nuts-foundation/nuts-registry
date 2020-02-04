#!/bin/sh

rm valid_files.tar.gz
cd valid_files
tar cvzf valid_files.tar.gz events/
cd ..
mv valid_files/*.tar.gz ./