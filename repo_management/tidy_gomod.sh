initial_dir=$(pwd)
modules=$(find ".." -type d -name 'pnp*')
root_module="github.com/go-pnp/go-pnp"

for i in $modules; do
  if [[ $i == *"../.git"* ]] || [[ $i == *"../_refactor"* ]]; then
    continue
  fi


  dir=${i#../}

  echo "Tidying ${dir}"
  cd "../${dir}"
  go mod tidy
  cd "${initial_dir}"

done
