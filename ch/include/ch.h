
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

#pragma omp
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
 * Performing the bucket scans is a crucial part of the many-to-many
 * computation.
 * It can be switched off to allow time measurements of the forward search
 * without accounting for the bucket scans.
 */
const bool performBucketScans = true;
const std::string PROFILEDIR = "/var/lib/spp/ch/";

const EdgeWeight Weight::MAX_VALUE;

Counter counter;

inline NodeID mapNodeID(const MyGraph* const g, const NodeID u) {
  // map the actual node ID to the node ID that is used internally by
  // highway-node routing
  return g->mapExtToIntNodeID(u);
}

MyGraph* loadGraph(const std::string& country, const int speed_profile,
                   const std::string& dataDir) {
  std::ostringstream filename;
  filename << dataDir << country << "-" << speed_profile << ".sgr";

  std::string path = filename.str();

  ifstream inGraph(path.c_str(), ios::binary);

  if (!inGraph) {
    std::cerr << "Input file '" << path << "' could not be read." << endl;
    exit(-1);
  }

  MyGraph* graph = new MyGraph(inGraph);
  inGraph.close();
  return graph;
}

const std::string calc_dm(const std::string& json_data,
                          const std::string& country, const int speed_profile,
                          const std::string& dataDir) {
  rapidjson::Document d;
  LevelID earlyStopLevel = 10;

  MyGraph* graph = loadGraph(country, speed_profile, dataDir);

  d.Parse<0>(json_data.c_str());

  const rapidjson::Value& sources = d["sources"];
  assert(sources.IsArray());
  vector<NodeID> v_sources;

  for (rapidjson::SizeType i = 0; i < sources.Size(); i++) {
    v_sources.push_back(mapNodeID(graph, (NodeID)sources[i].GetUint()));
  }

  ManyToMany<MyGraph, DijkstraManyToManyFW, DijkstraManyToManyBW,
             performBucketScans> mtm(graph, earlyStopLevel);
  Matrix<EdgeWeight> matrix((NodeID)sources.Size(), (NodeID)sources.Size());
  mtm.computeMatrix(v_sources, v_sources, matrix);

  int noOfRows = matrix.noOfRows();
  int noOfCols = matrix.noOfCols();

  rapidjson::Document out_doc;
  out_doc.SetObject();
  vector<string> names;

  for (int j = 0; j < noOfRows; j++) {
    rapidjson::Value result;
    result.SetArray();
    rapidjson::Document::AllocatorType& allocator = out_doc.GetAllocator();

    for (int k = 0; k < noOfCols; k++) {
      result.PushBack(matrix.value(j, k), allocator);
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

  delete graph;

  return strbuf.GetString();
}

const std::string calc_dm_par(const std::string& json_data,
                              const std::string& country,
                              const int speed_profile,
                              const std::string& dataDir) {
  std::cout << "parallel is enabled" << std::endl;
  rapidjson::Document d;
  LevelID earlyStopLevel = 10;

  std::stringstream filename;

  MyGraph* graph = loadGraph(country, speed_profile, dataDir);

  d.Parse<0>(json_data.c_str());

  const rapidjson::Value& sources = d["sources"];
  assert(sources.IsArray());
  vector<NodeID> v_sources;

  for (rapidjson::SizeType i = 0; i < sources.Size(); i++) {
    v_sources.push_back(mapNodeID(graph, (NodeID)sources[i].GetUint()));
  }
  rapidjson::Document out_doc;
  out_doc.SetObject();
  int max_thr = omp_get_max_threads();
  int thr_size = v_sources.size() / max_thr;
  int size_rest = v_sources.size() % max_thr;
  vector<int> matr_cols;
  for (int j = 0; j < max_thr; j++) {
    matr_cols.push_back(thr_size);
  }
  matr_cols[matr_cols.size() - 1] += size_rest;

#pragma omp parallel for ordered schedule(dynamic)
  for (int matr_id = 0; matr_id < max_thr; matr_id++) {
    MyGraph* const graph_thr = new MyGraph(*graph);
    int num_threads = omp_get_num_threads();
    ManyToMany<MyGraph, DijkstraManyToManyFW, DijkstraManyToManyBW,
               performBucketScans> mtm(graph_thr, earlyStopLevel);
    Matrix<EdgeWeight> matrix((NodeID)matr_cols[matr_id],
                              (NodeID)sources.Size());
    int offset = matr_id * thr_size;
    vector<NodeID>::const_iterator first = v_sources.begin() + offset;
    vector<NodeID>::const_iterator last =
        v_sources.begin() + offset + matr_cols[matr_id];
    vector<NodeID> newVec(first, last);

    mtm.computeMatrix(newVec, v_sources, matrix);

    int noOfRows = matrix.noOfRows();
    int noOfCols = matrix.noOfCols();

#pragma omp ordered
    {
      vector<string> names;
      for (int j = 0; j < noOfRows; j++) {
        rapidjson::Value result;
        result.SetArray();
        rapidjson::Document::AllocatorType& allocator = out_doc.GetAllocator();

        for (int k = 0; k < noOfCols; k++) {
          result.PushBack(matrix.value(j, k), allocator);
        }

        stringstream ss;
        ss << j + matr_id* thr_size;

        string str = ss.str();

        names.push_back(str);

        rapidjson::Value key;
        key.SetString(names[j].c_str(), names[j].length(),
                      out_doc.GetAllocator());
        out_doc.AddMember(key, result, allocator);
      }
    }
  }

  rapidjson::StringBuffer strbuf;
  rapidjson::Writer<rapidjson::StringBuffer> writer(strbuf);
  out_doc.Accept(writer);

  delete graph;

  return strbuf.GetString();
}

const std::string calc_paths(const std::string& json_data,
                             const std::string& country,
                             const int speed_profile,
                             const std::string& dataDir) {
  rapidjson::Document d;
  LevelID earlyStopLevel = 10;

  const clock_t begin_time = clock();

  MyGraph* graph = loadGraph(country, speed_profile, dataDir);

  d.Parse<0>(json_data.c_str());
  // cout <<"Load and parse: "<<float( clock () - begin_time ) /  CLOCKS_PER_SEC
  // <<endl;
  rapidjson::Document out_doc;
  out_doc.SetObject();

  rapidjson::Value result;
  result.SetArray();
  rapidjson::Document::AllocatorType& allocator = out_doc.GetAllocator();
  DijkstraManyToManyFW _dFW(graph);
  for (rapidjson::SizeType i = 0; i < d.Size(); i++) {
    // cout <<"loop begins: "<<float( clock () - begin_time ) /  CLOCKS_PER_SEC
    // <<endl;

    const rapidjson::Value& c = d[i];
    NodeID source_id = mapNodeID(graph, (NodeID)c["source"].GetUint());
    NodeID target_id = mapNodeID(graph, (NodeID)c["target"].GetUint());

    _dFW.clear();
    EdgeWeight w = _dFW.bidirSearch(source_id, target_id);
    Path a;
    _dFW.pathTo(a, target_id, -1, true, true);
    EdgeID num_edges = a.noOfEdges();

    rapidjson::Value result_internal;
    result_internal.SetArray();
    rapidjson::Value out_doc_internal;
    out_doc_internal.SetObject();

    for (EdgeID e = 0; e <= num_edges; e++) {
      result_internal.PushBack(a.node(e), out_doc.GetAllocator());
    }

    rapidjson::Value plen(w);
    out_doc_internal.AddMember("length", plen, out_doc.GetAllocator());
    out_doc_internal.AddMember("nodes", result_internal,
                               out_doc.GetAllocator());
    result.PushBack(out_doc_internal, out_doc.GetAllocator());
    // cout <<"loop end: "<<float( clock () - begin_time ) /  CLOCKS_PER_SEC
    // <<"Number: "<<num_edges<<endl;
  }

  out_doc.AddMember("edges", result, allocator);
  rapidjson::StringBuffer strbuf;
  rapidjson::Writer<rapidjson::StringBuffer> writer(strbuf);
  out_doc.Accept(writer);

  delete graph;

  return strbuf.GetString();
}
