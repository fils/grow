#!/bin/bash
# Note, this script leverage gron: https://github.com/tomnomnom/gron

ls_cmd() {
    rg -l "Corelyzer archive file" *.jsonld
}

# If you use this for ntriples, be sure to add in a graph in the URL target
for i in $(ls_cmd $1); do
   temp_file=$(mktemp)
  # temp2_file=$(mktemp)


   cat $i | jq \
   '.encodingFormat += ["application/zip", "https://opencoredata.org/voc/csdco/v1/Car"]' \
   >  $temp_file

    echo "----------------------------------"
    /bin/cat  $temp_file 

    cp $temp_file /home/fils/tmp/carpkgs/$i

    rm $temp_file

# cp $temp_file $i  # needed?  can jq do in place replacement?

done

