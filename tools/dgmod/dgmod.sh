#!/bin/bash
# Note, this script leverage gron: https://github.com/tomnomnom/gron

mc_cmd() {
        #mc ls $1 | awk '{print $5}'
        mc find $1   --name "*.jsonld"  
        # mc find $1   --name "*.jsonld"  | awk -F "/" '{print $5}'
        #echo "nas/doclouds/igsnsprint/bqs2e1iu6s73o70jh660.jsonld"
        
        # echo "ocdmsp/opencore/csdco/res/YUFL.jsonld"
        # echo "ocdmsp/opencore/csdco/do/00033daef38e12d329be17df28f5032b8adebb3bc4d98c08e0cc013c155603d7.jsonld"
        # echo "ocdmsp/opencore/csdco/pkg/000f0b65a1cf2cda4f038e7474660f1d8653040c86b7d2c1c32a1675697beace.jsonld"
        #i echo "ocdmsp/ocdtest/csdco/pkg/000f0b65a1cf2cda4f038e7474660f1d8653040c86b7d2c1c32a1675697beace.jsonld"
}

# If you use this for ntriples, be sure to add in a graph in the URL target
for i in $(mc_cmd $1); do
    #echo $i

    # get our JSON-LD file base 
    IFS='/' read -r -a sa <<< $i
    # strip off the .jsonld
    IFS="." read -r -a fa <<< ${sa[-1]}
    #echo ${fa[0]}

    #mc cat $1/$i | curl -X PUT --header "Content-Type:application/n-quads" -d @- http://localhost:3030/$2/data
    temp_file=$(mktemp)

# Note, SED can not do look ahead, look behind regex, so for that I use perl

# TYPE Resource 
    # mc cat $i |  gron | sed - \
    # -e 's/http:\/\/schema.org/https:\/\/schema.org/g'  \
    # -e 's/http:\/\/opencoredata.org/https:\/\/opencoredata.org/g' \
    # -e 's/opencoredata.org\/id\/do\//opencoredata.org\/id\/csdco\/res\//g' \
    # | gron  -u > $temp_file
    
# TYPE Digital Document
    mc cat $i |  gron | sed - \
    -e 's/http:\/\/schema.org/https:\/\/schema.org/g'  \
    -e 's/http:\/\/opencoredata.org/https:\/\/opencoredata.org/g' \
    -e 's/opencoredata.org\/id\/do\//opencoredata.org\/id\/csdco\/do\//g' \
    | /usr/bin/perl -pe "s/(?<=\/do\/)(.*)(?=\")/${fa[0]}/g"  \
    | gron  -u > $temp_file

# TYPE Dataset   Q:  Should the document ID end with .jsonld  (I do content neg, but perhaps I should still to not confuse with the DO)
    # mc cat $i |  gron | sed - \
    # -e 's/http:\/\/schema.org/https:\/\/schema.org/g'  \
    # -e 's/http:\/\/opencoredata.org/https:\/\/opencoredata.org/g' \
    # -e 's/opencoredata.org\/id\/do\//opencoredata.org\/id\/csdco\/do\//g' \
    # | /usr/bin/perl -pe "s/(?<=json.url = \"https:\/\/opencoredata\.org\/id\/csdco\/)(.*)(?=\")/pkg\/${fa[0]}/g"  \
    # | /usr/bin/perl -pe "s/(?<=json\[\"\@id\"\] = \"https:\/\/opencoredata\.org\/id\/csdco\/)(.*)(?=\")/pkg\/${fa[0]}/g"  \
    #  | /usr/bin/perl -pe "s/(?<=json.about = \"https:\/\/opencoredata\.org\/id\/csdco\/)(.*)(?=\/.*\")/pkg/g"  \
    # | gron  -u > $temp_file

    # convert location to spatial converage
    #mc cat $i |  gron | sed - -e 's/\.location/\.spatialCoverage/g' | gron  -u > $temp_file

    mc cp $temp_file $i
    #/usr/bin/bat -p   $temp_file 
    # /bin/cat  $temp_file 
    var=$(/usr/bin/wc -l  $temp_file)
    echo "$var for $i"

    rm $temp_file
done

