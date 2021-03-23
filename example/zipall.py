import glob
import os
for i in glob.glob("*"):
    print(i)
    if os.path.isdir(i) and i != "zipped":
        os.chdir(i)
        os.system("zip {} *".format(i))
        os.system("mv {}.zip ../zipped/{}".format(i, i))
        os.chdir("..")
