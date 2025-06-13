import tkinter as tk
from tkinter import filedialog, messagebox
from cryptography.fernet import Fernet
import mysql.connector
import os

# 글로벌 변수 선언
db_연결 = None
db_커서 = None

# 데이터베이스 연결 함수
def 데이터베이스에_연결(호스트, 사용자, 비밀번호, 데이터베이스):
    global db_연결, db_커서
    try:
        db_연결 = mysql.connector.connect(
            host=호스트,
            user=사용자,
            password=비밀번호,
            database=데이터베이스
        )
        db_커서 = db_연결.cursor()
        messagebox.showinfo("성공", "데이터베이스에 성공적으로 연결되었습니다.")
        # 사용자 ID 입력 필드, 파일 경로 선택 버튼, 암호화 및 복호화 버튼 프레임 생성
        생성_프레임()
    except mysql.connector.Error as err:
        messagebox.showerror("에러", "DB 정보가 맞지 않습니다.")

# 프레임 생성 함수
def 생성_프레임():
    # 파일 경로 선택 버튼
    파일_경로_라벨 = tk.Label(root, text="경로:")
    파일_경로_라벨.grid(row=2, column=0, padx=5, pady=5)
    global 파일_경로_입력
    파일_경로_입력 = tk.Entry(root)
    파일_경로_입력.grid(row=2, column=1, padx=5, pady=5)
    파일_찾기_버튼 = tk.Button(root, text="파일", 
                         command=lambda: 파일_경로_입력.insert(tk.END, filedialog.askopenfilename()))
    파일_찾기_버튼.grid(row=2, column=2, padx=5, pady=5)

    # 폴더 찾기 버튼 추가
    폴더_찾기_버튼 = tk.Button(root, text="폴더", 
                         command=lambda: 파일_경로_입력.insert(tk.END, filedialog.askdirectory()))
    폴더_찾기_버튼.grid(row=2, column=3, padx=5, pady=5)

    # 암호화 및 복호화 버튼
    암호화_버튼 = tk.Button(root, text="암호화", command=파일_암호화)
    암호화_버튼.grid(row=3, column=0, padx=5, pady=5)
    복호화_버튼 = tk.Button(root, text="복호화", command=파일_복호화)
    복호화_버튼.grid(row=3, column=1, padx=5, pady=5)

# 파일 암호화 함수
def 파일_암호화():
    # 암호화에 사용할 키 생성
    키 = 암호화_키_생성()
    
    # 파일 경로 또는 폴더 경로
    경로 = 파일_경로_입력.get()

    # 경로가 파일인지 디렉토리인지 확인
    if os.path.isfile(경로):
        파일_암호화_처리(경로, 키)
    elif os.path.isdir(경로):
        for root_dir, _, files in os.walk(경로):
            for file in files:
                파일_암호화_처리(os.path.join(root_dir, file), 키)
    else:
        messagebox.showerror("에러", "유효한 파일 또는 폴더 경로를 입력하세요.")
        return

    messagebox.showinfo("성공", "파일이 성공적으로 암호화되었습니다.")
    파일_경로_입력.delete(0, tk.END)  # 파일 경로 입력 박스 비우기

def 파일_암호화_처리(파일_경로, 키):
    # 파일 경로 및 키를 데이터베이스에 저장
    암호화_키_저장(파일_경로, 키)
    
    fernet = Fernet(키)
    
    # 파일 암호화
    with open(파일_경로, 'rb') as 파일:
        원본_데이터 = 파일.read()
    암호화된_데이터 = fernet.encrypt(원본_데이터)
    with open(파일_경로, 'wb') as 파일:
        파일.write(암호화된_데이터)
    
    # 암호화 로그 저장
    로그_저장(파일_경로, "encryption activated")

# 파일 복호화 함수
def 파일_복호화():
    # 파일 경로 또는 폴더 경로
    경로 = 파일_경로_입력.get()

    # 경로가 파일인지 디렉토리인지 확인
    if os.path.isfile(경로):
        파일_복호화_처리(경로) 
    elif os.path.isdir(경로):
        for root_dir, _, files in os.walk(경로):
            for file in files:
                파일_복호화_처리(os.path.join(root_dir, file))      
    else:
        messagebox.showerror("에러", "유효한 파일 또는 폴더 경로를 입력하세요.")
        return

    파일_경로_입력.delete(0, tk.END)  # 파일 경로 입력 박스 비우기

def 파일_복호화_처리(파일_경로):
    # 파일 경로를 데이터베이스에서 검색하여 암호화 키를 가져오기
    암호화된_키 = 암호화_키_가져오기(파일_경로)
    if not 암호화된_키:
        messagebox.showerror("에러", f"{파일_경로}에 대한 암호화 키를 찾을 수 없습니다.")
        return
    else:
        messagebox.showinfo("성공", f"{파일_경로}가 성공적으로 복호화되었습니다.")  
    
    fernet = Fernet(암호화된_키)
    
    # 파일 복호화
    with open(파일_경로, 'rb') as 파일:
        암호화된_데이터 = 파일.read()
    복호화된_데이터 = fernet.decrypt(암호화된_데이터)
    with open(파일_경로, 'wb') as 파일:
        파일.write(복호화된_데이터)

    암호화_키_삭제(파일_경로)

    # 복호화 로그 저장
    로그_저장(파일_경로, "decryption activated")

# 사용자마다의 고유한 암호화 키 생성 및 저장
def 암호화_키_생성():
    return Fernet.generate_key()

# 사용자마다의 암호화 키를 데이터베이스에 저장
def 암호화_키_저장(파일_경로, 키):
    db_커서.execute("INSERT INTO DLtable (filepath, EKey) VALUES (%s, %s)", (파일_경로, 키))
    db_연결.commit()

# 사용자마다의 암호화 키를 데이터베이스에서 가져오기
def 암호화_키_가져오기(파일_경로):
    db_커서.execute("SELECT EKey FROM DLtable WHERE filepath = %s", (파일_경로,))
    암호화된_키 = db_커서.fetchone()
    if 암호화된_키:
        return 암호화된_키[0]
    else:
        return None
    
# 사용자마다의 암호화 키를 데이터베이스에서 삭제
def 암호화_키_삭제(파일_경로):
    db_커서.execute("DELETE FROM DLtable WHERE filepath = %s", (파일_경로,))
    db_연결.commit()

# 로그 저장 함수
def 로그_저장(파일_경로, 작업_종류):
    db_커서.execute("INSERT INTO dl_log (filepath, operation) VALUES (%s, %s)", (파일_경로, 작업_종류))
    db_연결.commit()

# TKinter GUI 설정
root = tk.Tk()
root.title("DataLocker")

# 데이터베이스 연결 필드
db_프레임 = tk.Frame(root)
db_프레임.grid(row=0, columnspan=4, padx=5, pady=5)

호스트_라벨 = tk.Label(db_프레임, text="DB 호스트:")
호스트_라벨.grid(row=1, column=0, padx=5, pady=5)
호스트_입력 = tk.Entry(db_프레임)
호스트_입력.grid(row=1, column=1, padx=5, pady=5)

사용자_라벨 = tk.Label(db_프레임, text="DB 사용자:")
사용자_라벨.grid(row=2, column=0, padx=5, pady=5)
사용자_입력 = tk.Entry(db_프레임)
사용자_입력.grid(row=2, column=1, padx=5, pady=5)

비밀번호_라벨 = tk.Label(db_프레임, text="DB 비밀번호:")
비밀번호_라벨.grid(row=3, column=0, padx=5, pady=5)
비밀번호_입력 = tk.Entry(db_프레임, show="*")
비밀번호_입력.grid(row=3, column=1, padx=5, pady=5)

데이터베이스_라벨 = tk.Label(db_프레임, text="DB 이름:")
데이터베이스_라벨.grid(row=4, column=0, padx=5, pady=5)
데이터베이스_입력 = tk.Entry(db_프레임)
데이터베이스_입력.grid(row=4, column=1, padx=5, pady=5)

연결_버튼 = tk.Button(db_프레임, text="DB에 연결", command=lambda: 데이터베이스에_연결(호스트_입력.get(), 사용자_입력.get(), 비밀번호_입력.get(), 데이터베이스_입력.get()))
연결_버튼.grid(row=5, columnspan=2, padx=5, pady=5)

root.mainloop()
