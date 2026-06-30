import os

def replace_in_file(filepath, old_str, new_str):
    with open(filepath, 'r') as f:
        content = f.read()
    new_content = content.replace(old_str, new_str)
    if new_content != content:
        with open(filepath, 'w') as f:
            f.write(new_content)
        print(f"Updated {filepath}")

for root, dirs, files in os.walk('.'):
    # skip .git and hidden dirs
    if '/.' in root or root.startswith('./.'): continue
    for file in files:
        if file.endswith('.go') or file.endswith('.md') or file == 'go.mod':
            filepath = os.path.join(root, file)
            replace_in_file(filepath, 'psgo', 'fynescope')
            replace_in_file(filepath, 'Psgo', 'Fynescope')
            replace_in_file(filepath, 'PSGO', 'FYNESCOPE')

# Rename the test file
if os.path.exists('psgo_test.go'):
    os.rename('psgo_test.go', 'fynescope_test.go')
