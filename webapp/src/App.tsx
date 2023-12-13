import { useEffect, useState } from 'react';
import List from '@mui/material/List';
import ListItemText from '@mui/material/ListItemText';
import ListItemIcon from '@mui/material/ListItemIcon';
import FolderOpenOutlinedIcon from '@mui/icons-material/FolderOpenOutlined';
import InsertDriveFileOutlinedIcon from '@mui/icons-material/InsertDriveFileOutlined';
import ListItem from '@mui/material/ListItem';
import IconButton from '@mui/material/IconButton';
import IosShareOutlinedIcon from '@mui/icons-material/IosShareOutlined';
import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Divider from '@mui/material/Divider';
import Toolbar from '@mui/material/Toolbar';
import Grid from '@mui/material/Grid';

import QRCode from 'qrcode.react';

import DialogTitle from '@mui/material/DialogTitle';
import Dialog from '@mui/material/Dialog';

import { ListItemButton } from '@mui/material';

const DemoPaper = styled(Paper)(({ theme }) => ({
  width: "80%",
  padding: theme.spacing(2),
  ...theme.typography.body2,
  textAlign: 'center',
}));

interface File {
  file_name: string;
  file_modtime: string;
  is_dir: boolean;
  file_size: string; // Update this type based on your actual data
  sub_file_num: number;
  sub_dir_num: number;
}

function App() {
  const [files, setFiles] = useState<File[]>([]);
  const [path, setPath] = useState<string>();
  const [dialogState, setDialogState] = useState<boolean>(false);
  const [shareUrl, setShareUrl] = useState<string>('');
  const [localIp, setLocalIp] = useState("");
  const req = function (path: string) {
    fetch(`/api/files?path=${path}`)
      .then(response => response.json())
      .then(data => {
        if (data.files) {
          setFiles(data.files)
          setLocalIp(data.local_ip)
          if (path === undefined) {
            setPath(data.path)
          }
        } else {
          console.warn(data.message)
        }
      })
      .catch(error => console.error('Error:', error));
  }

  useEffect(() => {
    req(path!)
  }, [path]);

  const clickBtn = function (f: string, isDir: boolean, files: number) {
    if (!isDir) {
      const fileUrl = `/api/download?fname=${path}/${f}`;

      // 创建一个虚拟<a>标签
      const link = document.createElement('a');
      link.href = fileUrl;

      // 设置下载属性
      link.download = 'yourFileName.pdf';

      // 模拟点击
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      return
    }
    console.log(f)
    let newPath = ''
    if (path !== "/") {
      newPath = `${path}/${f}`
    } else {
      newPath = `/${f}`
    }

    setPath(newPath)
    if (files === 0) {
      setFiles([])
    }
    console.log("切换路径至：", newPath)
  }

  const share = function (fname: string) {
    if (path !== '/') {
      setShareUrl(`http://${localIp}/api/download?fname=${path}/${fname}`)
    } else {
      setShareUrl(`http://${localIp}/api/download?fname=/${fname}`)
    }
    setDialogState(true);
  }

  const clickBackBtn = function () {
    console.log("back..")
    if (path === undefined) {
      setPath('')
      return
    }
    const index = path?.lastIndexOf('/');

    if (index > 0) {
      const result = path?.substring(0, index);
      console.log(result);
      setPath(result)
      console.log("切换路径至：", result)
    } else if (index === 0) {
      setPath('/')
      console.log("切换路径至：", '/')
    } else {
      console.error("String does not contain a slash.");
    }

  }
  return (
    <Container maxWidth="lg">
      <Box style={{
        display: 'flex',
        minWidth: "650px",
        minHeight: "100vh",
        flexDirection: "column",
        alignItems: "center",
      }}>
        <Typography variant="h4" gutterBottom style={{ margin: "30px 70px", alignSelf: "flex-start" }}>
          Index of {path}
        </Typography>
        <DemoPaper elevation={6} square={false}>
          <List dense={false} >
            <ListItem key={-1} secondaryAction={<IconButton edge="end" aria-label="more"></IconButton>} disablePadding>
              <ListItemButton onClick={clickBackBtn}>
                <ListItemIcon>
                  <FolderOpenOutlinedIcon />
                </ListItemIcon>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <ListItemText primary={'../'} />
                  </Grid>
                  <Grid item xs={4}>
                  </Grid>
                  <Grid item xs={2}>
                  </Grid>
                </Grid>
              </ListItemButton>
            </ListItem>
            {files.map((item, index) => (
              [<Divider />,
              <ListItem key={index} secondaryAction={
                <IconButton onClick={() => { share(item.file_name) }} edge="end" aria-label="more">
                  <IosShareOutlinedIcon style={{display: item.is_dir ? 'none' : 'block'}}/>
                </IconButton>} disablePadding>
                <ListItemButton onClick={() => { clickBtn(item.file_name, item.is_dir, item.sub_dir_num + item.sub_file_num) }}>
                  <ListItemIcon>
                    {item.is_dir ? <FolderOpenOutlinedIcon /> : <InsertDriveFileOutlinedIcon />}
                  </ListItemIcon>
                  <Grid container spacing={2}>
                    <Grid item xs={6}>
                      <ListItemText primary={item.file_name} />
                    </Grid>
                    <Grid item xs={4}>
                      <ListItemText secondary={item.file_modtime} />
                    </Grid>
                    <Grid item xs={2}>
                      <ListItemText secondary={item.is_dir ? `${item.sub_dir_num} dirs ${item.sub_file_num} files` : item.file_size} />
                    </Grid>
                  </Grid>
                </ListItemButton>
              </ListItem>]
            ))}
          </List>
        </DemoPaper>
        <Toolbar style={{ flexShrink: 0 }}>
          <Typography variant="body1" color="inherit">
            © File Share Tool. By <a href={`mailto:tumble-leap@outlook.com`} onClick={() => { window.location.href = `mailto:tumble-leap@outlook.com` }}>Mr. Chen</a>
          </Typography>
        </Toolbar>
        <Dialog open={dialogState} onClose={() => { setDialogState(false) }}>
          <DialogTitle><QRCode value={shareUrl} /></DialogTitle>
        </Dialog>
      </Box>
    </Container>

  );
}

export default App;
