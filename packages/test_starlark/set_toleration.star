def run(r, spec):
  for resource in r:
    if spec.get("filter") != None:
      if resource["metadata"]["name"] != spec["filter"]["name"]:
        continue
    # mutate the resource
    resource["spec"]["template"]["spec"]["tolerations"] = spec["tolerations"]

# get the value of the annotation to add
spec = ctx.resource_list["functionConfig"]["spec"]

run(ctx.resource_list["items"], spec)
