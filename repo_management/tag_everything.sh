read -p "This will tag all modules. Are you sure? (y/n) " -n 1 -r
echo "\n"
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  exit 1
fi

read -p "Version: " version

# if version is empty, exit
if [ -z "$version" ]; then
  echo "Version cannot be empty"
  exit 1
fi

# if version doesn't have semver format,
if [[ ! $version =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Version must be in semver format"
  exit 1
fi

# if version doesn't start with 'v', add it
if [[ ! $version =~ ^v ]]; then
  version="v${version}"
fi

echo "All modules will be tagged with version ${version}"

for i in $(find "." -type d -name 'pnp*'); do
  # skip .git and _refactor directory
  if [[ $i == *".git"* ]] || [[ $i == *"_refactor"* ]]; then
    continue
  fi
  # Remove leading './' using parameter expansion
  dir=${i#./}
  module="github.com/go-pnp/go-pnp/${dir}"

  echo "Tagging with version ${dir}/${version}"
  git tag -a ${dir}/${version} -m "Module ${dir} version ${version}"
done

git push --tags
