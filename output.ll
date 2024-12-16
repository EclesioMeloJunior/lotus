; ModuleID = 'main'
source_filename = "main"

%String = type { i32, ptr }

define void @print(%String %0) {
entry:
  ret void
}

define i32 @test() {
entry:
  ret i32 2
}

define i32 @main() {
entry:
  %a = alloca i32, align 4
  store i32 3, ptr %a, align 4
  %b = alloca i32, align 4
  store i32 0, ptr %b, align 4
  %c = alloca i32, align 4
  %b1 = load i32, ptr %b, align 4
  store i32 %b1, ptr %c, align 4
  %name = alloca %String, align 8
  store [8 x i8] c"Eclesio\00", ptr %name, align 1
  %a2 = alloca i32, align 4
  store i32 1, ptr %a2, align 4
  %calltest = call i32 @test()
  ret i32 %calltest
}
