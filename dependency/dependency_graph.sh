#!/bin/bash
terraform graph | dot -Tpng > ../img/dependeny_graph.png
