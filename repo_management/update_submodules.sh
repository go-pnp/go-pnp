initial_dir=$(pwd)
modules=$(find ".." -type d -name 'pnp*')
root_module="github.com/go-pnp/go-pnp"

for i in $modules; do
  if [[ $i == *"../.git"* ]] || [[ $i == *"../_refactor"* ]]; then
    echo "Skipping $i"
    continue
  fi


  dir=${i#../}
  module="github.com/go-pnp/go-pnp/${dir}"

  cd "../${dir}" || continue

  # Initialize the module if go.mod does not exist
  if [ ! -f go.mod ]; then
    echo "Initializing ${module}"
    go mod init "${module}"
  fi
  echo "Updating dependencies for ${module}"

  go get -u github.com/go-pnp/go-pnp@v0.0.2
  go get -u
  # Update each dependency to its latest version
#  for dep in $modules; do
#
#    if [[ $i == *"../.git"* ]] || [[ $i == *"../_refactor"* ]]; then
#      echo "Skipping dependency $i"
#      continue
#    fi
#
#    mod_dir=${dep#../}
#    # List tags, filter for the current module, sort them, and pick the latest
#    latest_version=$(git tag -l "${mod_dir}/v*" | sort -V | tail -n1 | sed "s#${mod_dir}/v##")
#    if [ ! -z "$latest_version" ]; then
#      echo "`pwd` Updating ${root_module}/${mod_dir}@v${latest_version} for ${dir} "
#      go get "${root_module}/${mod_dir}@v${latest_version}"
#    fi
#
#
#  done

  go mod tidy
  git add go.mod go.sum
  cd "${initial_dir}"
done
