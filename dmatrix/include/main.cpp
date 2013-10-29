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
std::ostringstream ss;
std::cout<<omp_get_max_threads();

for(int i = 0; i < 5000; i++)
{
    ss.str("");
    ss<<randomNodeID(noOfNodes);
    jsonInput +=  ss.str();
    jsonInput += ",";
}
jsonInput[jsonInput.size()-1] = ']';
jsonInput = "{\"sources\":[" + jsonInput + "}";

calc((char *)jsonInput.c_str(),"germany",40);

calc_par((char *)jsonInput.c_str(),"germany",40);

return 0;
}
