
#include <rapidjson/include/rapidjson/document.h>
#include <rapidjson/include/rapidjson/prettywriter.h>
#include <rapidjson/include/rapidjson/stringbuffer.h>
#include <rapidjson/include/rapidjson/writer.h>

#include <iostream>
#include <string>
#include <iomanip>
#include <fstream>
#include <sstream>

using namespace std;

#include "config.h"
#include "stats/utils.h"
#include "datastr/graph/graph.h"
#include "datastr/graph/SearchGraph.h"
#include "processing/DijkstraCH.h"

#include "manyToMany.h"

typedef datastr::graph::SearchGraph TransitGraph;
typedef datastr::graph::SearchGraph MyGraph;
typedef DijkstraCHManyToManyFW DijkstraManyToManyFW;
typedef DijkstraCHManyToManyBW DijkstraManyToManyBW;

/*
 * Performing the bucket scans is a crucial part of the many-to-many computation.
 * It can be switched off to allow time measurements of the forward search
 * without accounting for the bucket scans.
 */
const bool performBucketScans = true;
const std::string PROFILEDIR = "/var/lib/spp/ch/";

const EdgeWeight Weight::MAX_VALUE;

Counter counter;


inline NodeID mapNodeID(const MyGraph *const g, const NodeID u) {
    // map the actual node ID to the node ID that is used internally by highway-node routing
    return g->mapExtToIntNodeID(u);
}

//extern "C" {
//    const char* calc(char* json_data);
//}

MyGraph* loadGraph(const string& filename)
{
    ifstream inGraph( filename.c_str(), ios::binary );
    if (! inGraph) {
        cerr << "Input file '" << filename << "' not found." << endl;
        exit(-1);
    }
    
    MyGraph* graph = new MyGraph(inGraph);
    inGraph.close();
    return graph;
}

const char* calc(char* json_data, const char* country, const int speed_profile)
{
    rapidjson::Document d;
    LevelID earlyStopLevel = 10;

    std::stringstream filename;

    filename << PROFILEDIR << country << "-" << speed_profile << ".sgr"; 

    MyGraph* graph = loadGraph(filename.str());

    d.Parse<0>(json_data);
    
    const rapidjson::Value& sources = d["sources"]; 
    assert(sources.IsArray());
    vector<NodeID> v_sources;
    
    for (rapidjson::SizeType i = 0; i < sources.Size(); i++)
    {
        v_sources.push_back(mapNodeID(graph, (NodeID)sources[i].GetUint()));
    }

    ManyToMany<MyGraph, DijkstraManyToManyFW, DijkstraManyToManyBW, performBucketScans> mtm(graph, earlyStopLevel);
    Matrix<EdgeWeight> matrix((NodeID)sources.Size(), (NodeID)sources.Size());
    mtm.computeMatrix(v_sources, v_sources, matrix);

    int noOfRows = matrix.noOfRows();
    int noOfCols = matrix.noOfCols();

    rapidjson::Document out_doc;
    out_doc.SetObject();
    vector<string> names;

    for (int j = 0; j < noOfRows; j++)
    {
        rapidjson::Value result;
        result.SetArray();
        rapidjson::Document::AllocatorType& allocator = out_doc.GetAllocator();

        for (int k = 0; k < noOfCols; k++)
        {
            result.PushBack(matrix.value(j,k), allocator);
        }

        stringstream ss;
        ss << j;
        string str = ss.str();
        names.push_back(str);
        out_doc.AddMember(names[j].c_str(), result, out_doc.GetAllocator());
    }

    rapidjson::StringBuffer strbuf;
    rapidjson::Writer<rapidjson::StringBuffer> writer(strbuf);
    out_doc.Accept(writer);

    string res = strbuf.GetString();
    int len = res.length();
    char* result = (char*)malloc(len);
    strncpy(result, res.c_str(), len);
    strcat(result, "\0");

    delete graph;
    
    return result;
}

