# STEP3: アルゴリズムとデータ構造　

STEP3 では、基本的なアルゴリズムとデータ構造を学んだ後、LeetCode の問題を用いて演習を行います。解説の前に問題を解いておくことをお勧めします。

## 教材

**:book: Reference**

- (JA) [現役シリコンバレーエンジニアが教えるアルゴリズム・データ構造・コーディングテスト入門](https://mercari.udemy.com/course/python-algo/)

- (EN) [Python Data Structures & Algorithms + LEETCODE Exercises](https://mercari.udemy.com/course/data-structures-algorithms-python/)

**:beginner: point**

- まず以下の言葉について調べ、どのようなものであるか説明できるようになりましょう
  - 時間計算量と空間計算量
  - ビッグオー記法
  - 連想配列
- 以下の基本的なアルゴリズムについて Udemy 等で勉強し、解説できるようになりましょう。
  - バイナリサーチとは、どのようなアルゴリズムですか？バイナリサーチの計算量が $O(\log n)$ である理由を説明してください。
  - LinkedList と Array の違いについて説明してください。
  - ハッシュテーブルを説明し、計算量を見積もってください。
  - グラフ探索アルゴリズムについて説明し、BFS (Breadth First Search) や DFS (Depth First Search) の使い分けについて説明してください。

## 演習

### [Word Pattern](https://leetcode.com/problems/word-pattern/description/)

英小文字からなるパターン `p` と、空白区切りの文字列 `s` が与えられるので、`s` が `p` に従うかどうかを判定してください。 例えば、`p = "abba"`, `s = "dog cat cat dog"` の場合 `s` は `p` に従い、`p="abba"`, `s="dog cat cat fish"` の場合 `s` は `p` に従いません。

**:beginner: checkpoint**

#### Step1: 文字列 `s` を空白で区切る方法を考えてみましょう。

<details>
<summary>ヒント</summary>

- 各言語では、文字列操作のためのライブラリや関数などが標準で提供されているはずです
- Web 検索や ChatGPT を駆使して、"文字列 空白区切り" などで検索してみましょう
</details>

#### Step2: パターン `p` の書く文字が、`s` のどの部分に対応するかを管理する方法を考えてみましょう。

<details>
<summary>ヒント</summary>

- 例えば、Example 1 の場合、`p` の各文字に対応する `s` 内の単語は、`a => dog`, `b => cat` です
- このような対応を管理するために、辞書やハッシュテーブルを使うと良いでしょう
- 例えば、Python では、`dict` を使って、`p` の各文字に対応する `s` 内の単語を管理できます
- こちらも、Web 検索や ChatGPT を駆使して、"Python 辞書" などで検索してみましょう
</details>

### 実行コード

```python
class Solution(object):
    def wordPattern(self, pattern, s):
        """
        :type pattern: str
        :type s: str
        :rtype: bool
        """

        a = ""
        b = ""
        c = ""
        result = ""
        a_ = ""
        b_ = ""
        result_ =""
        words = s.split()
        for i in range(len(words)):
            word = words[i]
            if i == 0: # 一週目
                a = word
                result += "a"
                a_ = word
                result_ += "b"
            else: # 二週目
                if word != a: # iがaでなければ
                    b = word
                    result += "b"
                    result_ += "a"
                elif word != a and word != a_ and word != b and word != b_:
                    result += "c"
                    result_ += "c"
                else:
                    result += "a"
                    result_ += "b"
        if pattern == result or pattern == result_:
            return True
        else:
            return False


```

### [Find All Numbers Disappeared in an Array](https://leetcode.com/problems/find-all-numbers-disappeared-in-an-array/description/)

n 個の整数からなる配列 nums が与えられ、nums[i] は [1, n] の範囲にあります。この配列に現れない [1, n] の範囲のすべての整数を返してください。

**:beginner: checkpoint**

#### Step1: O(n^2)-time and O(1)-space で解く

<details>
<summary>ヒント</summary>

- シンプルなな 2 重ループを用いて、O(n^2)-time and O(1)-space で解けます
</details>

#### Step2: O(n)-time and O(n)-space で解く

<details>
<summary>ヒント</summary>

- 配列 nums 内に要素が出現したかどうかを記録するための配列を用意することで、O(n)-time and O(n)-space で解けます
</details>

#### 発展: O(n)-time and O(1)-space で解く (おまけ)

入力と返り値を除いて、O(1)-space で解くことは可能ですか？

<details>
<summary>ヒント</summary>

- 深く考察をすると、O(n)-time and O(1)-space で解けることがわかります
- 解説で扱う予定なので、挑戦してみてください
</details>

### 実行コード

```python
class Solution:
    def findDisappearedNumbers(self, nums: List[int]) -> List[int]:
        result = []
        for i in range(1, len(nums)+1): # numsの長さ分だけ回す
            if i not in nums:
                result.append(i)
        return result
```

```python
class Solution:
    def findDisappearedNumbers(self, nums: List[int]) -> List[int]:
        # 各数字が存在することを nums 内で印をつけて記録
        for i in range(len(nums)):
            index = abs(nums[i]) - 1  # 対応するインデックス
            nums[index] = -abs(nums[index])  # マーク（負の数にする）

        result = []
        for i in range(len(nums)):
            if nums[i] > 0:
                result.append(i + 1)  # 出現しなかった数字

        return result
```

### [Intersection of Two Linked Lists](https://leetcode.com/problems/intersection-of-two-linked-lists/description)

2 つの単方向 Linked List が与えられるので、2 つのリストが交差するノードを返してください。交差しない場合は、`null` を返してください。

**:beginner: checkpoint**

#### Step1: O(n)-time and O(n)-space で解く

<details>
<summary>ヒント</summary>

- Hash Table を使ってノードを記録することで、O(n)-time and O(n)-space で解けます
</details>

#### Step2: O(n)-time and O(1)-space で解く

入力と返り値を除いて、O(1)-space で解くことは可能ですか？

<details>
<summary>ヒント</summary>

- 2 つのリストの長さを比較して、長いリストを短いリストと同じ長さにすることで、O(n)-time and O(1)-space で解けます
- 解説で扱う予定です
</details>

### 実行コード

```python
class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        def getLength(head):
            length = 0
            while head:
                length += 1
                head = head.next
            return length

        lenA = getLength(headA)
        lenB = getLength(headB)

        # ポインタをセット
        currA = headA
        currB = headB

        # 長い方を進めて長さを合わせる
        if lenA > lenB:
            for _ in range(lenA - lenB):
                currA = currA.next
        else:
            for _ in range(lenB - lenA):
                currB = currB.next

        # 同時に進めて一致するノードを探す
        while currA and currB:
            if currA == currB:
                return currA
            currA = currA.next
            currB = currB.next

        return None
```

```python
class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        # 2つのポインタを使う
        pA = headA
        pB = headB

        # 2つのポインタが一致するまで繰り返す
        while pA != pB:
            # それぞれのポインタが終端に達したら、もう一方のリストの先頭に移動
            pA = pA.next if pA else headB
            pB = pB.next if pB else headA

        # 一致したノード（またはNone）を返す
        return pA
```

```python
class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        # ポインタAとポインタBを、それぞれのリストの先頭にセットする
        pointerA = headA
        pointerB = headB

        # ポインタAとポインタBが同じノードを指すまでループする
        while pointerA != pointerB:
            # ポインタAがNoneでなければ次のノードに進む
            if pointerA is not None:
                pointerA = pointerA.next
            else:
                # ポインタAがNoneになったら、Bの先頭に切り替える
                pointerA = headB

            # ポインタBがNoneでなければ次のノードに進む
            if pointerB is not None:
                pointerB = pointerB.next
            else:
                # ポインタBがNoneになったら、Aの先頭に切り替える
                pointerB = headA

        # 最後に一致したノード（またはNone）を返す
        return pointerA

```

#### 発展: two pointers を使って解く方法 (おまけ)

<details>
<summary>ヒント</summary>

- 片方の tail から head にポインタをはり、Floyd's Linked List Cycle Finding Algorithm に帰着する
</details>

### [Koko Eating Bananas](https://leetcode.com/problems/koko-eating-bananas/) (optional)

### [Non-overlapping Intervals](https://leetcode.com/problems/non-overlapping-intervals/description/) (optional)

### [Longest Substring Without Repeating Characters](https://leetcode.com/problems/longest-substring-without-repeating-characters/description/) (optional)

[STEP4: 出品 API を作る](./04-api.ja.md)
