via - fast shortest path computation
====================================

**via** is a new, lightweight shortest path problem computation service fully implemented using RESTful techniques. via provides distance matrix calculation through a simple API. via makes use of the fantastic [OpenStreetMap](http://www.openstreetmap.org) map data for its computation.

**Note**: currently, via relies on precomputed contraction hierarchies graphs which are built on OpenStreetMap data. They are essential for via to work.  While we have published the source code for via, we currently have not released the preprocessed graph files. We intend to release *instructions* on how to do this but due to size and bandwidth constraints, we have not yet found a feasible solution for publishing these files. Instructions will be released soon(tm).

Installing via
--------------

To compile via, first install the following:

  * rapidson (https://code.google.com/p/rapidjson/)
  * Go development version from source (https://code.google.com/p/go/)
  * SWIG 2.0.11

Once those are taken care of, simply run:

    go get github.com/co-sky-developers/via/

Then copy the ``production.json`` or ``development.json`` configuration files and simply call it by running ``via``.

Performance
-----------

On an Amazon EC2 large instance with two cores, current benchmarks give about 5 seconds for a 1000x1000 distance matrix using Germany's road network, which is the largest and densest in Europe. 

Documentation
-------------

API documentation will be available soon.

Acknowledgements
----------------

> If I have seen further it is by standing on the shoulders of giants.
>   *- Isaac Newton*

via includes source code of the contraction hierarchies project, see [here](http://algo2.iti.kit.edu/routeplanning.php) for more information, which is also licensed under the AGPL.

Requirements
------------

* Go development version
* Redis for fast key-value storage
* Swig 2.0.11 for C++ wrapping
* PostgreSQL 
* Preprocessed OpenStreetMap data using contraction hierarchies. *We do not currently provide these, but instructions on how to compile them yourself will be given in the future.*

About
-----

via is a project by the CO-SKY research team.

http://www.co-sky.fi

License
-------

via is redistributed using the GNU Affero General Public License, version 3 (19th Nov 2007). See LICENSE for further instructions for redistribution.
