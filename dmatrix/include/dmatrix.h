
#include <rapidjson/document.h>
#include <rapidjson/prettywriter.h>
#include <rapidjson/stringbuffer.h>
#include <rapidjson/writer.h>

#include <iostream>
#include <string>
#include <iomanip>
#include <fstream>
#include <sstream>

#include <omp.h>
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

#pragma GCC optimize("openmp")

const char* calc_par(char* json_data, const char* country, const int speed_profile)
{
    std::cout << "parallel is enabled" << std::endl;
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
    rapidjson::Document out_doc;
    out_doc.SetObject();
    int max_thr = omp_get_max_threads();
    int thr_size = v_sources.size()/max_thr;
    int size_rest = v_sources.size() % max_thr;
    vector<int> matr_cols;
    for (int j = 0; j < max_thr; j++)
    {
        matr_cols.push_back(thr_size);
    }
    matr_cols[matr_cols.size()-1] += size_rest;
    

    
    
    
    #pragma omp parallel for ordered schedule(dynamic)
    for(int matr_id = 0; matr_id < max_thr; matr_id++)
    {
    MyGraph *const graph_thr = new MyGraph(*graph);
    int num_threads = omp_get_num_threads();
    ManyToMany<MyGraph, DijkstraManyToManyFW, DijkstraManyToManyBW, performBucketScans> mtm(graph_thr, earlyStopLevel);
    Matrix<EdgeWeight> matrix((NodeID)matr_cols[matr_id], (NodeID)sources.Size());
    int offset = matr_id*thr_size;
    vector<NodeID>::const_iterator first = v_sources.begin() + offset;
    vector<NodeID>::const_iterator last = v_sources.begin() + offset + matr_cols[matr_id];
    vector<NodeID> newVec(first, last);
    
    mtm.computeMatrix(newVec, v_sources, matrix);

    int noOfRows = matrix.noOfRows();
    int noOfCols = matrix.noOfCols();

    
    
    #pragma omp ordered
    {
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
        ss << j + matr_id*thr_size;
        
        string str = ss.str();

        names.push_back(str);

        rapidjson::Value key;
        key.SetString(names[j].c_str(), names[j].length(), out_doc.GetAllocator());
        out_doc.AddMember(key, result, allocator);
        
    }

    }

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



