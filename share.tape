# Where should we write the GIF?
Output demo.gif

# Set up a 1200x600 terminal with 46px font.
Set Width 1000
Set Height 850

Set Shell zsh
Sleep 500ms
Type "echo Lorem Ipsum >"
Sleep 0.5s
Type "testfile.txt"
Enter
Sleep 500ms
Type "share testfile.txt"
Enter
Sleep 4.5s
Escape
Type "OA dropbox"
Sleep 500ms
Enter
Sleep 5s
Type "share --setup"
Enter
Escape
Type "[I"
Sleep 500ms
Enter
Escape
Type "[I"
Down
Sleep 3.5s
Escape
Type "[A"
Sleep 1s
Enter
Sleep 500ms
Escape
Type "OA"
Sleep 1s
Ctrl+W
Ctrl+W
Type "share testfile.txt"
Enter
Sleep 3.5s
Type "share --setup"
Enter
Escape
Type "[I"
Sleep 1s
Enter
Escape
Type "[I"
Down 2
Enter
Sleep 2.5s
Ctrl+D

