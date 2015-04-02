mkdir -p $root
mkdir -p $root/$path_subtree
git clone $bare_subtree $root/$path_subtree
mkdir -p $root/$path_dep
git clone $bare_dep $root/$path_dep
(cd $root/$path_dep && git reset --hard HEAD~1)
