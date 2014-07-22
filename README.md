via - fast shortest path computation
====================================

**via** is a new, lightweight shortest path problem computation service fully implemented using RESTful techniques. via provides distance matrix calculation through a simple API. via makes use of the fantastic [OpenStreetMap](http://www.openstreetmap.org) map data for its computation.

**Note**: currently, via relies on precomputed contraction hierarchies graphs which are built on OpenStreetMap data. They are essential for via to work.  While we have published the source code for via, we currently have not released the preprocessed graph files. We intend to release *instructions* on how to do this but due to size and bandwidth constraints, we have not yet found a feasible solution for publishing these files. Instructions will be released soon(tm).

Installing via
--------------

To compile via, first install the following (Debian/Ubuntu packages in parens):

  * rapidjson 0.11 (https://code.google.com/p/rapidjson/)
  * boost C++ libraries (libboost-dev)
  * Go 1.3 (use godeb: http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go)
  * Redis
  * SWIG 3.0.2

Once those are taken care of, simply run:

    go get -u github.com/nfleet/via/

Then copy the ``config_template.json`` configuration files, modify it accordingly, and simply call it by running ``via <config_file>``. Once you've established that via works, you need to figure out a way to send contraction hierarchies node data to the service. 

Performance
-----------

On an Amazon EC2 large instance with two cores, current benchmarks give about 5 seconds for a 1000x1000 distance matrix using the German road network, which is the largest and densest in Europe (ca. 100 million graph nodes).

Acknowledgements
----------------

> If I have seen further it is by standing on the shoulders of giants.
>   *- Isaac Newton*

via includes source code of the contraction hierarchies project, see [here](http://algo2.iti.kit.edu/routeplanning.php) for more information, which is also licensed under the AGPL.

About
-----

via is a project by [NFleet](http://www.nfleet.fi).

License
-------

via is redistributed using the GNU Affero General Public License, version 3 (19th Nov 2007). See LICENSE for further instructions for redistribution.
