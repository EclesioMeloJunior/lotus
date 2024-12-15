; ModuleID = 'main'
source_filename = "main"

%String = type { i32, ptr }

define void @print(%String %0) {
entry:
  ret void
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
  %a3 = load i32, ptr %a2, align 4
  %b4 = load i32, ptr %b, align 4
  %addtmp = add i32 %a3, %b4
  ret i32 %addtmp
}

define void @test() {
entry:
  %a = alloca i32, align 4
  store i32 2, ptr %a, align 4
  ret void
}
