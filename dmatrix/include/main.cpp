#include <sstream>
#include "dmatrix.h"
#include <iostream>
#include <omp.h>

inline NodeID randomNodeID(NodeID n) {
    return (NodeID)(rand() / (double)(RAND_MAX+1.0) * n);
}

int main()
{
srand(time(NULL));
const NodeID noOfNodes = 10589551;
std::string jsonInput = "";
std::string jsonInput_s = "";
std::string jsonInput_t = "";
std::ostringstream ss;
ss.str("");
ss<<randomNodeID(noOfNodes);
jsonInput_s = ss.str();
ss.str("");
ss<<randomNodeID(noOfNodes);
jsonInput_t = ss.str();


jsonInput = "{\"source\":" + jsonInput_s + ",\"target\":" + jsonInput_t + "}";
cout<<jsonInput<<endl;
EdgeWeight path_len;
EdgeID num_edges;
cout<<calc_path((char *)jsonInput.c_str(),"germany",40);


}
