# Sources

## Intro

While GROW is focused on a _document_ based approach that doesn't mean that it 
cant not be used in with approach approaches.  The following sections will touch 
on two of these; structured data files and relational databases (SQL).

These approaches are NOT in GROW at this time.  However, they were known directions 
and as such elements of the code are already set up to accommodate their development 
should there be support or reason to.  

> NOTE:  The following approaches are for the case where a person does _not_ want 
> to materialize rows or views into digital objects (files).  Rather, it is more about
> how these approaches could be connected to GROW

## Structured data files

Here we will call "structured data" serializations like CSV, JSON or Parquet.  The 
GROW architecture is based on the Amazon S3 API.  This also means it can be extended
to scope the  _Select API_. The open source Minio package has support for the S3 Select API.  

> Implementing this approach in Google Cloud would like involve Big Query with external data sources support for 
> Google Cloud Storage.  The author is not aware of current support for the S3 Select API in Google Cloud Storage. 

S3 Select would allow us to pull a row from one of these structured data files.  


Current, GROW performs an object lookup for a URL like:


| URL                                                     | Object (Bucket +  Prefix)                  |
|---------------------------------------------------------|--------------------------------------------|
| ```https://example.org/id/artifact/gaurdianofforever``` | ```/mybucket/artifact/gaurdianofforever``` |

To support data expressed in tabular data it would now perform an S3 Select call on a tabular data object:

| URL                                                     | Select API example                  |
|---------------------------------------------------------|--------------------------------------------|
| ```https://example.org/id/artifact/gaurdianofforever``` | ```select * from s3object s where s.ID like 'gaurdianofforever' ``` |

In this way, items (rows) in a tabular data file could be mapped to a URL.

Note, that there would need to be logic to convert the row values to some desired representation for the net.
This would be something like JSON-LD or GeoJSON.  A JSON-LD representation would be needed at a minimum to facilitate 
discovery.  For the use case of generation for GROW this would be a type schame.org/Dataset, it could be anything though.
Care would need to be taken to ensure the tabular data contains the elements necessary to populate such a schema. 


## Relational DBs

In a case where the data is held in relational database a virtual dataset based on something
like Frictionless Data Packages could be generated.  The "data" though would not be a pointer to a file
but rather a pointer to the SQL call (or REST call).    The package would contain the pointers to the data
and an associated JSON-LD document would hold the discovery data.  This approach needs more details to better 
describe it.  

* http://wiki.esipfed.org/index.php/Interagency_Data_Stewardship/Citations/provider_guidelines 
* https://commons.esipfed.org/node/7980 