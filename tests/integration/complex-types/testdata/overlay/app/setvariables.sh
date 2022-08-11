#!/bin/bash

echo corral_set number=1
echo corral_set singlequotednumber='1'
echo corral_set doublequotednumber="1"

echo corral_set string=abc
echo corral_set singlequotedstring='abc'
echo corral_set doublequotedstring="abc"

echo corral_set array=[1,2,3]
echo corral_set singlequotedarray='[1,2,3]'
echo corral_set doublequotedarray="[1,2,3]"

echo corral_set object=\{\"a\":1,\"b\":2.0,\"c\":\"3\",\"d\":[4,\"5\"]\}
echo corral_set singlequotedobject='{"a":1,"b":2.0, "c":"3", "d":[4, "5"]}'
echo corral_set doublequotedobject="{\"a\":1,\"b\":2.0, \"c\":\"3\", \"d\":[4, \"5\"]}"

echo corral_set string_output="a"
echo corral_set number_output=1
echo corral_set array_output="[1,2,3]"
echo corral_set object_output='{"a":1,"b":2.0, "c":"3", "d":[4, "5"]}'
