import time

print("Training model...")
time.sleep(5)

with open("output.txt", "w") as f:
    f.write("Demo experiment completed successfully.\n")

with open("stop.txt", "w") as f:
    f.write("done\n")
