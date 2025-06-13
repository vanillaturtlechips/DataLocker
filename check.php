<?php
if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $dbid = $_POST['dbid'];
    $password = $_POST['password'];

    $servername = "localhost";
    $dbname = "dl";

    $password2 = strtoupper(hash('sha256', $password));

    $conn = new mysqli($servername, $dbid, $password2, $dbname);

    if ($conn->connect_error) {
        die("DB 연결 실패");
    }

    $sql = "SELECT * FROM dl_log";
    $result = $conn->query($sql);

    if ($result->num_rows > 0) {
        echo "<h2>DataLocker Log:</h2>";
        echo "<table border='1'>
                <tr>
                    <th>filepath</th>
		    <th>operation</th>
		    <th>timestamp</th>
                </tr>";
        while($row = $result->fetch_assoc()) {
            echo "<tr>
                    <td>" . $row["filepath"]. "</td>
                    <td>" . $row["operation"]. "</td>
                    <td>" . $row["timestamp"]. "</td>
                  </tr>";
        }
        echo "</table>";
    } else {
        echo "0 results";
    }
    $conn->close();
} else {
    echo "잘못된 요청";
}
?>

